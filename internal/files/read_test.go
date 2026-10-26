package files_test

import (
	"aws-cli-guard/internal/files"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRegularMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "missing.toml")

	_, err := files.ReadRegular(path)
	assertErrorIs(t, err, os.ErrNotExist)
}

func TestReadRegularSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.toml")
	writeRegularFile(t, target, []byte(fixtureSecret+"-hidden"), 0o600)

	link := filepath.Join(dir, "config.toml")
	mustSymlink(t, target, link)

	_, err := files.ReadRegular(link)
	assertErrorIs(t, err, files.ErrSymlink)
	assertNoSecretInError(t, err)
}

func TestReadRegularDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	_, err := files.ReadRegular(dir)
	assertErrorIs(t, err, files.ErrNotRegular)
}

func TestReadRegularLoosePermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "loose.toml")
	body := []byte(fixtureSecret + "-readable")
	writeRegularFile(t, path, body, 0o644)

	got := mustReadRegular(t, path)
	if string(got) != string(body) {
		t.Fatalf("got %q, want %q", got, body)
	}
}
