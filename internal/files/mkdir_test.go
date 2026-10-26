package files_test

import (
	"aws-cli-guard/internal/files"
	"path/filepath"
	"testing"
)

func TestMkdirPrivateCreatesLeaf0700(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	ancestor := filepath.Join(base, "config")
	mustMkdir(t, ancestor, 0o755)

	leaf := filepath.Join(ancestor, "aws-cli-guard")
	mustMkdirPrivate(t, leaf)

	if fileModeBits(t, ancestor) != 0o755 {
		t.Fatalf("ancestor should stay 0755, got %o", fileModeBits(t, ancestor))
	}

	if fileModeBits(t, leaf) != 0o700 {
		t.Fatalf("leaf want 0700, got %o", fileModeBits(t, leaf))
	}
}

func TestMkdirPrivateExistingDirectoryChmodsLeaf(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	leaf := filepath.Join(dir, "state")
	mustMkdir(t, leaf, 0o755)

	mustMkdirPrivate(t, leaf)

	if fileModeBits(t, leaf) != 0o700 {
		t.Fatalf("leaf want 0700, got %o", fileModeBits(t, leaf))
	}
}

func TestMkdirPrivateSymlinkLeaf(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	realDir := filepath.Join(base, "real")
	mustMkdir(t, realDir, 0o755)

	link := filepath.Join(base, "link")
	mustSymlink(t, realDir, link)

	err := files.MkdirPrivate(link)
	assertErrorIs(t, err, files.ErrSymlink)

	if fileModeBits(t, realDir) != 0o755 {
		t.Fatalf("target mode changed")
	}
}
