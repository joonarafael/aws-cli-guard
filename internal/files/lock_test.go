package files_test

import (
	"aws-cli-guard/internal/files"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpenLockCreatesLockFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	grants := filepath.Join(dir, "grants.json")
	lock := mustOpenLock(t, grants)

	lockPath := grants + ".lock"
	if fileModeBits(t, lockPath) != 0o600 {
		t.Fatalf("lock mode want 0600, got %o", fileModeBits(t, lockPath))
	}

	mustRelease(t, lock)

	_, statErr := os.Stat(lockPath)
	if statErr != nil {
		t.Fatalf("lock file should remain: %v", statErr)
	}

	lock2 := mustOpenLock(t, grants)
	mustRelease(t, lock2)
}

func TestOpenLockMutualExclusion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	grants := filepath.Join(dir, "grants.json")
	assertLockBlocksUntilRelease(t, grants)
}

func assertLockBlocksUntilRelease(t *testing.T, grantsPath string) {
	t.Helper()

	holding := make(chan struct{})

	go holdLockBriefly(t, grantsPath, holding)

	<-holding

	var acquiredEarly atomic.Bool

	done := make(chan struct{})

	go waitForLock(t, grantsPath, &acquiredEarly, done)

	time.Sleep(100 * time.Millisecond)

	if acquiredEarly.Load() {
		t.Fatal("second lock acquired while first still held")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("second lock did not acquire after first release")
	}
}

func holdLockBriefly(t *testing.T, grantsPath string, holding chan struct{}) {
	t.Helper()

	first, err := files.OpenLock(grantsPath)
	if err != nil {
		t.Errorf("first OpenLock: %v", err)
		close(holding)

		return
	}

	close(holding)
	time.Sleep(300 * time.Millisecond)

	releaseErr := first.Release()
	if releaseErr != nil {
		t.Errorf("first Release: %v", releaseErr)
	}
}

func waitForLock(t *testing.T, grantsPath string, acquiredEarly *atomic.Bool, done chan struct{}) {
	t.Helper()

	second, err := files.OpenLock(grantsPath)
	if err != nil {
		t.Errorf("second OpenLock: %v", err)
		close(done)

		return
	}

	acquiredEarly.Store(true)

	releaseErr := second.Release()
	if releaseErr != nil {
		t.Errorf("second Release: %v", releaseErr)
	}

	close(done)
}

func TestOpenLockSymlinkLockFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustMkdirPrivate(t, dir)

	target := filepath.Join(dir, "elsewhere.lock")
	writeRegularFile(t, target, nil, 0o600)

	grants := filepath.Join(dir, "grants.json")
	lockLink := grants + ".lock"
	mustSymlink(t, target, lockLink)

	_, err := files.OpenLock(grants)
	assertErrorIs(t, err, files.ErrSymlink)
}

func TestOpenLockMissingParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	grants := filepath.Join(dir, "missing", "grants.json")

	_, err := files.OpenLock(grants)
	assertErrorIs(t, err, os.ErrNotExist)
}
