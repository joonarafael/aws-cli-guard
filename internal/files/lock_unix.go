//go:build unix

package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

// Lock holds an exclusive advisory lock on path.lock until Release.
type Lock struct {
	mu       sync.Mutex
	path     string
	fd       int
	released bool
}

// OpenLock opens path.lock and acquires an exclusive flock (blocking).
func OpenLock(path string) (*Lock, error) {
	lockPath := path + ".lock"

	err := parentDirExists(lockPath)
	if err != nil {
		return nil, err
	}

	lockFD, err := openLockFile(lockPath)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(lockFD, syscall.LOCK_EX)
	if err != nil {
		_ = syscall.Close(lockFD)

		return nil, fmt.Errorf("aws-cli-guard: flock %q: %w", lockPath, err)
	}

	return &Lock{mu: sync.Mutex{}, path: lockPath, fd: lockFD, released: false}, nil
}

// Release unlocks and closes the lock file, leaving it on disk.
func (l *Lock) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.released {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrLockAlreadyReleased, l.path)
	}

	err := syscall.Flock(l.fd, syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: funlock %q: %w", l.path, err)
	}

	err = syscall.Close(l.fd)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: close lock %q: %w", l.path, err)
	}

	l.released = true

	return nil
}

func parentDirExists(path string) error {
	dir := filepath.Dir(path)

	info, err := os.Lstat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("aws-cli-guard: open lock %q: %w", path, os.ErrNotExist)
		}

		return fmt.Errorf("aws-cli-guard: stat parent of %q: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, dir)
	}

	if !info.IsDir() {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrNotDirectory, dir)
	}

	return nil
}

func openLockFile(lockPath string) (int, error) {
	lockFD, err := syscall.Open(
		lockPath,
		syscall.O_RDWR|syscall.O_CREAT|syscall.O_NOFOLLOW,
		fileMode,
	)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return 0, fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, lockPath)
		}

		if errors.Is(err, syscall.ENOENT) {
			return 0, fmt.Errorf("aws-cli-guard: open lock %q: %w", lockPath, os.ErrNotExist)
		}

		return 0, fmt.Errorf("aws-cli-guard: open lock %q: %w", lockPath, err)
	}

	err = os.Chmod(lockPath, fileMode)
	if err != nil {
		_ = syscall.Close(lockFD)

		return 0, fmt.Errorf("aws-cli-guard: chmod lock %q: %w", lockPath, err)
	}

	return lockFD, nil
}
