package cache

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Recover moves live cache entries aside and ensures an empty cache directory.
// It does not prompt. ErrAbsent means the directory was missing and nothing was created.
func Recover(env []string, now time.Time) ([]string, error) {
	path, err := paths.CacheDir(env)
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: cache: resolve path: %w", err)
	}

	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAbsent
		}

		return nil, fmt.Errorf("aws-cli-guard: cache: stat: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		aside, moveErr := files.MoveAside(path, now)
		if moveErr != nil {
			return nil, mapRecoverError(path, moveErr)
		}

		err = files.MkdirPrivate(path)
		if err != nil {
			return []string{aside}, mapRecoverError(path, err)
		}

		return []string{aside}, nil
	}

	asides, err := moveLiveChildren(path, now)
	if err != nil {
		return asides, err
	}

	err = files.MkdirPrivate(path)
	if err != nil {
		return asides, mapRecoverError(path, err)
	}

	return asides, nil
}

// Present reports whether the cache directory exists. ErrAbsent means it is missing.
func Present(env []string) error {
	path, err := paths.CacheDir(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: cache: resolve path: %w", err)
	}

	_, err = os.Lstat(path)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return ErrAbsent
	}

	return fmt.Errorf("aws-cli-guard: cache: stat: %w", err)
}

func moveLiveChildren(dir string, now time.Time) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: cache: read directory: %w", err)
	}

	var asides []string

	for _, entry := range entries {
		name := entry.Name()
		if isPreservedName(name) {
			continue
		}

		child := filepath.Join(dir, name)

		aside, moveErr := files.MoveAside(child, now)
		if moveErr != nil {
			return asides, mapRecoverError(child, moveErr)
		}

		asides = append(asides, aside)
	}

	return asides, nil
}

func isPreservedName(name string) bool {
	return strings.Contains(name, ".corrupt.") || strings.Contains(name, ".partial.")
}

func mapRecoverError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) ||
		errors.Is(err, files.ErrNotDirectory) {
		return fmt.Errorf("aws-cli-guard: cache: recover %q: %w", path, err)
	}

	return fmt.Errorf("aws-cli-guard: cache: recover %q: %w", path, err)
}
