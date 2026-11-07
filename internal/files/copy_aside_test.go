package files_test

import (
	"aws-cli-guard/internal/files"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCopyAsideRegularFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "config.toml")
	body := []byte(fixtureSecret + "-live")
	writeRegularFile(t, live, body, 0o600)

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	aside, err := files.CopyAside(live, now)
	if err != nil {
		t.Fatalf("CopyAside: %v", err)
	}

	gotLive, readErr := os.ReadFile(live)
	if readErr != nil {
		t.Fatalf("ReadFile live: %v", readErr)
	}

	if string(gotLive) != string(body) {
		t.Fatalf("live file should be unchanged")
	}

	gotAside, readErr := os.ReadFile(aside)
	if readErr != nil {
		t.Fatalf("ReadFile aside: %v", readErr)
	}

	if string(gotAside) != string(body) {
		t.Fatalf("aside bytes mismatch")
	}

	wantSuffix := ".migrated.20260115T120000.000000000Z"
	if !stringsHasSuffix(aside, wantSuffix) {
		t.Fatalf("aside = %q want suffix %q", aside, wantSuffix)
	}
}

func TestCopyAsideSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	target := filepath.Join(dir, "target.toml")
	writeRegularFile(t, target, []byte("target"), 0o600)

	link := filepath.Join(dir, "config.toml")
	mustSymlink(t, target, link)

	_, err := files.CopyAside(link, time.Now())
	if !errors.Is(err, files.ErrSymlink) {
		t.Fatalf("err = %v, want ErrSymlink", err)
	}
}

func TestCopyAsideMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.toml")

	_, err := files.CopyAside(missing, time.Now())
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want ErrNotExist", err)
	}

	assertNoSecretInError(t, err)
}

func TestCopyAsideTakenFirstCandidate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "grants.json")
	writeRegularFile(t, live, []byte("live"), 0o600)

	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	taken := live + ".migrated.20260115T120000.000000000Z"
	writeRegularFile(t, taken, []byte("taken"), 0o600)

	aside, err := files.CopyAside(live, now)
	if err != nil {
		t.Fatalf("CopyAside: %v", err)
	}

	want := live + ".migrated.20260115T120000.000000000Z.2"
	if aside != want {
		t.Fatalf("aside = %q want %q", aside, want)
	}
}
