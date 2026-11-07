//go:build unix

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// CopyAside copies path to a sibling .migrated.<utc-timestamp> name without removing the live file.
func CopyAside(path string, now time.Time) (string, error) {
	data, _, err := ReadRegular(path)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: copy aside %q: %w", path, err)
	}

	var lastErr error

	for _, aside := range migratedPathCandidates(path, now) {
		_, statErr := os.Lstat(aside)
		if statErr == nil {
			lastErr = fmt.Errorf("aws-cli-guard: aside name taken %q: %w", aside, syscall.EEXIST)

			continue
		}

		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("aws-cli-guard: stat aside %q: %w", aside, statErr)
		}

		writeErr := writeAsideFile(aside, data)
		if writeErr == nil {
			return aside, nil
		}

		if errors.Is(writeErr, syscall.EEXIST) {
			lastErr = writeErr

			continue
		}

		return "", writeErr
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("aws-cli-guard: %w: %q", ErrNoMigratedName, path)
}

func writeAsideFile(aside string, data []byte) error {
	partial, err := createPartialFile(aside, time.Now(), data)
	if err != nil {
		return err
	}

	err = os.Rename(partial, aside)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: rename %q to %q: %w", partial, aside, err)
	}

	return nil
}
