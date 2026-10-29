package policy

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Recover moves the policy file aside and writes the init default. It does not prompt.
func Recover(env []string, now time.Time) (string, error) {
	path, err := paths.ConfigFile(env)
	if err != nil {
		return "", mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return "", mapRecoverError(path, err)
	}

	_, err = os.Lstat(path)
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

	writeErr := files.WriteAtomic(path, []byte(DefaultFileBody))
	if writeErr != nil {
		return aside, mapRecoverError(path, writeErr)
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

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: policy: %w", err))
}
