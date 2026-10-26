//go:build unix

package files

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirMode = 0o700

// MkdirPrivate creates path and missing ancestors. Directories this call creates are 0700.
// The leaf is always 0700 when the call succeeds.
func MkdirPrivate(path string) error {
	clean := filepath.Clean(path)

	info, err := os.Lstat(clean)
	if err == nil {
		return finishExistingDir(clean, info)
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("aws-cli-guard: stat %q: %w", clean, err)
	}

	return createDirTree(clean)
}

func finishExistingDir(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, path)
	}

	if !info.IsDir() {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrNotDirectory, path)
	}

	err := os.Chmod(path, dirMode)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: chmod %q: %w", path, err)
	}

	return nil
}

func createDirTree(path string) error {
	parent := filepath.Dir(path)
	if parent != path {
		err := ensureParentDir(parent)
		if err != nil {
			return err
		}
	}

	err := os.Mkdir(path, dirMode)
	if err != nil {
		if os.IsExist(err) {
			info, statErr := os.Lstat(path)
			if statErr != nil {
				return fmt.Errorf("aws-cli-guard: stat %q: %w", path, statErr)
			}

			return finishExistingDir(path, info)
		}

		return fmt.Errorf("aws-cli-guard: mkdir %q: %w", path, err)
	}

	chmodErr := os.Chmod(path, dirMode)
	if chmodErr != nil {
		return fmt.Errorf("aws-cli-guard: chmod %q: %w", path, chmodErr)
	}

	return nil
}

func ensureParentDir(parent string) error {
	info, err := os.Lstat(parent)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, parent)
		}

		if !info.IsDir() {
			return fmt.Errorf("aws-cli-guard: %w: %q", ErrNotDirectory, parent)
		}

		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("aws-cli-guard: stat %q: %w", parent, err)
	}

	return createDirTree(parent)
}
