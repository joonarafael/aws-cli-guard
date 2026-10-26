package files_test

import (
	"aws-cli-guard/internal/files"
	"errors"
	"os"
	"strings"
	"testing"
)

func assertNoSecretInError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	if strings.Contains(err.Error(), fixtureSecret) {
		t.Fatalf("error must not contain fixture secret: %v", err)
	}
}

func assertErrorIs(t *testing.T, err error, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("errors.Is(%v, %v) = false, want true", err, target)
	}
}

func fileModeBits(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat %q: %v", path, err)
	}

	return info.Mode().Perm()
}

func writeRegularFile(t *testing.T, path string, data []byte, perm os.FileMode) {
	t.Helper()

	err := os.WriteFile(path, data, perm)
	if err != nil {
		t.Fatalf("WriteFile %q: %v", path, err)
	}
}

func mustMkdirPrivate(t *testing.T, path string) {
	t.Helper()

	err := files.MkdirPrivate(path)
	if err != nil {
		t.Fatalf("MkdirPrivate %q: %v", path, err)
	}
}

func mustWriteAtomic(t *testing.T, path string, data []byte) {
	t.Helper()

	err := files.WriteAtomic(path, data)
	if err != nil {
		t.Fatalf("WriteAtomic %q: %v", path, err)
	}
}

func mustReadRegular(t *testing.T, path string) []byte {
	t.Helper()

	data, err := files.ReadRegular(path)
	if err != nil {
		t.Fatalf("ReadRegular %q: %v", path, err)
	}

	return data
}

func mustOpenLock(t *testing.T, path string) *files.Lock {
	t.Helper()

	lock, err := files.OpenLock(path)
	if err != nil {
		t.Fatalf("OpenLock %q: %v", path, err)
	}

	return lock
}

func mustRelease(t *testing.T, lock *files.Lock) {
	t.Helper()

	err := lock.Release()
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()

	err := os.Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink %q -> %q: %v", link, target, err)
	}
}

func mustMkdir(t *testing.T, path string, perm os.FileMode) {
	t.Helper()

	err := os.Mkdir(path, perm)
	if err != nil {
		t.Fatalf("Mkdir %q: %v", path, err)
	}
}
