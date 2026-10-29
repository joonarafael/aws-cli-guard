package files_test

import (
	"aws-cli-guard/internal/files"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMoveAsideRegularFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "config.toml")
	body := []byte(fixtureSecret + "-live")
	writeRegularFile(t, live, body, 0o600)

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	aside, err := files.MoveAside(live, now)
	if err != nil {
		t.Fatalf("MoveAside: %v", err)
	}

	_, statErr := os.Stat(live)
	if !os.IsNotExist(statErr) {
		t.Fatalf("live path should be gone: %v", statErr)
	}

	got, readErr := os.ReadFile(aside)
	if readErr != nil {
		t.Fatalf("ReadFile aside: %v", readErr)
	}

	if string(got) != string(body) {
		t.Fatalf("aside bytes mismatch")
	}

	wantSuffix := ".corrupt.20260115T120000.000000000Z"
	if !stringsHasSuffix(aside, wantSuffix) {
		t.Fatalf("aside = %q want suffix %q", aside, wantSuffix)
	}
}

func TestMoveAsideSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	target := filepath.Join(dir, "target.toml")
	targetBody := []byte(fixtureSecret + "-target")
	writeRegularFile(t, target, targetBody, 0o600)

	link := filepath.Join(dir, "config.toml")
	mustSymlink(t, target, link)

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	aside, err := files.MoveAside(link, now)
	if err != nil {
		t.Fatalf("MoveAside: %v", err)
	}

	gotTarget, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("ReadFile target: %v", readErr)
	}

	if string(gotTarget) != string(targetBody) {
		t.Fatalf("target unchanged")
	}

	info, err := os.Lstat(aside)
	if err != nil {
		t.Fatalf("Lstat aside: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("aside should still be a symlink")
	}

	_, statErr := os.Stat(link)
	if !os.IsNotExist(statErr) {
		t.Fatal("live symlink should be gone")
	}
}

func TestMoveAsideDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	child := filepath.Join(dir, "cache")

	err := os.Mkdir(child, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	aside, err := files.MoveAside(child, now)
	if err != nil {
		t.Fatalf("MoveAside: %v", err)
	}

	info, err := os.Stat(aside)
	if err != nil {
		t.Fatalf("Stat aside: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("aside should be a directory")
	}

	_, statErr := os.Stat(child)
	if !os.IsNotExist(statErr) {
		t.Fatal("live path should be gone")
	}
}

func TestMoveAsideMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.toml")

	_, err := files.MoveAside(missing, time.Now())
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want ErrNotExist", err)
	}

	assertNoSecretInError(t, err)
}

func TestMoveAsideTakenFirstCandidate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "grants.json")
	writeRegularFile(t, live, []byte("live"), 0o600)

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	taken := live + ".corrupt.20260115T120000.000000000Z"
	writeRegularFile(t, taken, []byte("taken"), 0o600)

	aside, err := files.MoveAside(live, now)
	if err != nil {
		t.Fatalf("MoveAside: %v", err)
	}

	want := live + ".corrupt.20260115T120000.000000000Z.2"
	if aside != want {
		t.Fatalf("aside = %q want %q", aside, want)
	}

	got := mustReadRegular(t, aside)
	if string(got) != "live" {
		t.Fatalf("got %q", got)
	}
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
