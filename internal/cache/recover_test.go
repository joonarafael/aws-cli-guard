package cache_test

import (
	"aws-cli-guard/internal/cache"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixtureSecret = "AKIAIOSFODNN7EXAMPLE-cache-body"

func TestRecoverAbsentCreatesNothing(t *testing.T) {
	t.Parallel()

	env := cacheDirEnv(t)

	asides, err := cache.Recover(env, time.Now())
	if !errors.Is(err, cache.ErrAbsent) {
		t.Fatalf("err = %v, want ErrAbsent", err)
	}

	if asides != nil {
		t.Fatalf("asides = %v", asides)
	}

	_, err = os.Stat(cacheDirFromEnv(t, env))
	if !os.IsNotExist(err) {
		t.Fatal("cache dir should not be created")
	}
}

func TestRecoverMovesChildren(t *testing.T) {
	t.Parallel()

	env := cacheDirEnv(t)
	dir := cacheDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	live := filepath.Join(dir, "entry.bin")
	body := []byte(fixtureSecret + "-live")

	err = os.WriteFile(live, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	preserved := filepath.Join(dir, "old.partial.20200101T000000.000000000Z")

	err = os.WriteFile(preserved, []byte("keep"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	asides, err := cache.Recover(env, now)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	if len(asides) != 1 {
		t.Fatalf("asides = %v", asides)
	}

	_, statErr := os.Stat(live)
	if !os.IsNotExist(statErr) {
		t.Fatal("live entry should be moved")
	}

	got, err := os.ReadFile(preserved)
	if err != nil {
		t.Fatalf("ReadFile preserved: %v", err)
	}

	if string(got) != "keep" {
		t.Fatal("preserved partial should remain")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() == "entry.bin" {
			t.Fatal("live name should not remain")
		}
	}
}

func TestRecoverSymlinkDirNotFollowed(t *testing.T) {
	t.Parallel()

	env := cacheDirEnv(t)
	parent := filepath.Dir(cacheDirFromEnv(t, env))
	realCacheDir := filepath.Join(parent, "real-cache")
	targetFile := filepath.Join(realCacheDir, fixtureSecret+"-inside")

	err := os.MkdirAll(realCacheDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(targetFile, []byte("secret"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	link := cacheDirFromEnv(t, env)

	err = os.Symlink(realCacheDir, link)
	if err != nil {
		t.Fatal(err)
	}

	asides, err := cache.Recover(env, time.Now())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	if len(asides) != 1 {
		t.Fatalf("asides = %v", asides)
	}

	_, statErr := os.Stat(targetFile)
	if statErr != nil {
		t.Fatal("target inside real dir must be unchanged")
	}

	info, err := os.Stat(link)
	if err != nil {
		t.Fatalf("Stat new cache: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("live path should be empty directory")
	}
}

func TestPresentAbsent(t *testing.T) {
	t.Parallel()

	env := cacheDirEnv(t)

	err := cache.Present(env)
	if !errors.Is(err, cache.ErrAbsent) {
		t.Fatalf("err = %v", err)
	}
}

func cacheDirEnv(t *testing.T) []string {
	t.Helper()

	base := t.TempDir()

	return []string{"AWS_CLI_GUARD_CACHE_DIR=" + filepath.Join(base, "aws-cli-guard")}
}

func cacheDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_CACHE_DIR="
		if strings.HasPrefix(entry, prefix) {
			return entry[len(prefix):]
		}
	}

	t.Fatal("AWS_CLI_GUARD_CACHE_DIR not in env")

	return ""
}
