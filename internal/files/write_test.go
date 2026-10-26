package files_test

import (
	"aws-cli-guard/internal/files"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicNewFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "config.toml")
	body := []byte(fixtureSecret + "-live")
	mustWriteAtomic(t, live, body)

	got := mustReadRegular(t, live)
	if string(got) != string(body) {
		t.Fatalf("got %q, want %q", got, body)
	}

	if fileModeBits(t, live) != 0o600 {
		t.Fatalf("live file mode want 0600, got %o", fileModeBits(t, live))
	}

	partials, globErr := filepath.Glob(live + ".partial.*")
	if globErr != nil {
		t.Fatalf("Glob: %v", globErr)
	}

	if len(partials) != 0 {
		t.Fatalf("expected no partials after commit, got %v", partials)
	}
}

func TestWriteAtomicReplacesRegularFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "grants.json")
	writeRegularFile(t, live, []byte("old"), 0o600)

	next := []byte(fixtureSecret + "-updated")
	mustWriteAtomic(t, live, next)

	got := mustReadRegular(t, live)
	if string(got) != string(next) {
		t.Fatalf("got %q, want %q", got, next)
	}
}

func TestWriteAtomicKeepsSiblingPartial(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	live := filepath.Join(dir, "grants.json")
	sibling := live + ".partial.20200101T000000.000000000Z"
	siblingBody := []byte("partial-bytes-never-live")
	writeRegularFile(t, sibling, siblingBody, 0o600)

	liveBody := []byte(fixtureSecret + "-new-live")
	mustWriteAtomic(t, live, liveBody)

	gotSibling, readErr := os.ReadFile(sibling)
	if readErr != nil {
		t.Fatalf("ReadFile sibling: %v", readErr)
	}

	if string(gotSibling) != string(siblingBody) {
		t.Fatalf("sibling changed: got %q", gotSibling)
	}

	gotLive := mustReadRegular(t, live)
	if string(gotLive) != string(liveBody) {
		t.Fatalf("live got %q", gotLive)
	}
}

func TestWriteAtomicDestinationSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	target := filepath.Join(dir, "target.toml")
	targetBody := []byte(fixtureSecret + "-target")
	writeRegularFile(t, target, targetBody, 0o600)

	link := filepath.Join(dir, "config.toml")
	mustSymlink(t, target, link)

	err := files.WriteAtomic(link, []byte("attempt"))
	assertErrorIs(t, err, files.ErrSymlink)
	assertNoSecretInError(t, err)

	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("ReadFile target: %v", readErr)
	}

	if string(got) != string(targetBody) {
		t.Fatalf("target bytes changed")
	}

	partials, globErr := filepath.Glob(link + ".partial.*")
	if globErr != nil {
		t.Fatalf("Glob: %v", globErr)
	}

	if len(partials) != 0 {
		t.Fatalf("expected no partials, got %v", partials)
	}
}

func TestWriteAtomicDestinationDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	err := files.WriteAtomic(dir, []byte("x"))
	assertErrorIs(t, err, files.ErrNotRegular)
	assertNoSecretInError(t, err)

	partials, globErr := filepath.Glob(dir + ".partial.*")
	if globErr != nil {
		t.Fatalf("Glob: %v", globErr)
	}

	if len(partials) != 0 {
		t.Fatalf("expected no partials, got %v", partials)
	}
}

func TestWriteAtomicMissingParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	live := filepath.Join(dir, "missing", "parent", "file.toml")

	err := files.WriteAtomic(live, []byte("x"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}

	_, statErr := os.Stat(filepath.Join(dir, "missing"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected parent not created")
	}
}
