package grants

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Recover moves the grants file aside and writes an empty grant list. It does not prompt.
func Recover(env []string, now time.Time) (string, error) {
	path, err := paths.GrantsFile(env)
	if err != nil {
		return "", mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return "", mapRecoverError(path, err)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return "", mapRecoverError(path, err)
	}

	aside, recoverErr := recoverLocked(path, now)
	releaseErr := lock.Release()

	if recoverErr != nil {
		return aside, recoverErr
	}

	if releaseErr != nil {
		return aside, mapReleaseError(path, releaseErr)
	}

	return aside, nil
}

func recoverLocked(path string, now time.Time) (string, error) {
	_, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", newError(
				KindMissing,
				path,
				"",
				0,
				0,
				fmt.Errorf("%w: %w", ErrMissing, os.ErrNotExist),
			)
		}

		return "", mapRecoverError(path, err)
	}

	aside, err := files.MoveAside(path, now)
	if err != nil {
		return "", mapRecoverError(path, err)
	}

	writeErr := writeLocked(path, nil)
	if writeErr != nil {
		return aside, writeErr
	}

	return aside, nil
}

func mapRecoverError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	if _, ok := errors.AsType[*Error](err); ok {
		return err
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: grants: %w", err))
}
