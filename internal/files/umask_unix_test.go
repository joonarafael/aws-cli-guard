//go:build unix

package files_test

import (
	"path/filepath"
	"syscall"
	"testing"
)

//nolint:paralleltest // umask is process-wide; must not run parallel with other tests.
func TestWriteAtomicRespectsUmask(t *testing.T) {
	old := syscall.Umask(0)
	defer syscall.Umask(old)

	dir := t.TempDir()
	leaf := filepath.Join(dir, "state", "aws-cli-guard")
	mustMkdirPrivate(t, leaf)

	if fileModeBits(t, leaf) != 0o700 {
		t.Fatalf("dir mode want 0700, got %o", fileModeBits(t, leaf))
	}

	live := filepath.Join(leaf, "grants.json")
	mustWriteAtomic(t, live, []byte("g"))

	if fileModeBits(t, live) != 0o600 {
		t.Fatalf("file mode want 0600, got %o", fileModeBits(t, live))
	}
}
