//go:build unix

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// MoveAside renames path to a sibling .corrupt.<utc-timestamp> name without deleting data.
func MoveAside(path string, now time.Time) (string, error) {
	_, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("aws-cli-guard: move aside %q: %w", path, os.ErrNotExist)
		}

		return "", fmt.Errorf("aws-cli-guard: stat %q: %w", path, err)
	}

	var lastErr error

	for _, aside := range corruptPathCandidates(path, now) {
		_, statErr := os.Lstat(aside)
		if statErr == nil {
			lastErr = fmt.Errorf("aws-cli-guard: aside name taken %q: %w", aside, syscall.EEXIST)

			continue
		}

		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("aws-cli-guard: stat aside %q: %w", aside, statErr)
		}

		err = os.Rename(path, aside)
		if err == nil {
			return aside, nil
		}

		if renameAsideRetryable(err) {
			lastErr = err

			continue
		}

		return "", fmt.Errorf("aws-cli-guard: rename %q to %q: %w", path, aside, err)
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("aws-cli-guard: %w: %q", ErrNoCorruptName, path)
}

func renameAsideRetryable(err error) bool {
	return errors.Is(err, syscall.EEXIST)
}
