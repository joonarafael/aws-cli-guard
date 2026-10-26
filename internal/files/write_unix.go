//go:build unix

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

const fileMode = 0o600

// WriteAtomic replaces path with data using a same-directory partial file.
// A race that swaps in a symlink between the pre-rename Lstat and Rename is accepted;
// Rename does not follow a symlink into its target.
func WriteAtomic(path string, data []byte) error {
	err := lstatDestForWrite(path)
	if err != nil {
		return err
	}

	partial, err := createPartialFile(path, time.Now(), data)
	if err != nil {
		return err
	}

	err = lstatDestForWrite(path)
	if err != nil {
		return err
	}

	err = os.Rename(partial, path)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: rename %q to %q: %w", partial, path, err)
	}

	return nil
}

func lstatDestForWrite(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("aws-cli-guard: stat %q: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, path)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrNotRegular, path)
	}

	return nil
}

func createPartialFile(livePath string, now time.Time, data []byte) (string, error) {
	var lastErr error

	for _, partial := range partialPathCandidates(livePath, now) {
		partialFD, err := openPartialExclusive(partial)
		if err != nil {
			if partialNameRetryable(err) {
				lastErr = err

				continue
			}

			return "", err
		}

		writeErr := writeSyncClose(partialFD, partial, data)
		if writeErr != nil {
			return "", writeErr
		}

		return partial, nil
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("aws-cli-guard: %w: %q", ErrNoPartialName, livePath)
}

func partialNameRetryable(err error) bool {
	return errors.Is(err, syscall.EEXIST) || errors.Is(err, syscall.ELOOP)
}

func openPartialExclusive(partial string) (int, error) {
	partialFD, err := syscall.Open(
		partial,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW,
		fileMode,
	)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return 0, fmt.Errorf("aws-cli-guard: create partial %q: %w", partial, os.ErrNotExist)
		}

		if errors.Is(err, syscall.ELOOP) || errors.Is(err, syscall.EEXIST) {
			return 0, fmt.Errorf("aws-cli-guard: partial name taken %q: %w", partial, err)
		}

		return 0, fmt.Errorf("aws-cli-guard: create partial %q: %w", partial, err)
	}

	chmodErr := os.Chmod(partial, fileMode)
	if chmodErr != nil {
		_ = syscall.Close(partialFD)

		return 0, fmt.Errorf("aws-cli-guard: chmod partial %q: %w", partial, chmodErr)
	}

	return partialFD, nil
}

func writeSyncClose(partialFD int, partial string, data []byte) error {
	if len(data) > 0 {
		err := writeAll(partialFD, data)
		if err != nil {
			return fmt.Errorf("aws-cli-guard: write partial %q: %w", partial, err)
		}
	}

	err := syscall.Fsync(partialFD)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: sync partial %q: %w", partial, err)
	}

	err = syscall.Close(partialFD)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: close partial %q: %w", partial, err)
	}

	return nil
}

func writeAll(partialFD int, data []byte) error {
	for len(data) > 0 {
		count, err := syscall.Write(partialFD, data)
		if err != nil {
			return fmt.Errorf("syscall write: %w", err)
		}

		data = data[count:]
	}

	return nil
}
