package settings

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultFileBody is the init default for settings.toml (schema version 1).
const DefaultFileBody = `version = 1

[identity]
verify = false
`

// Create writes the default settings file when it is missing. It returns created true when
// a new file was written. An existing regular file is left unchanged (created false, nil error).
func Create(env []string) (bool, error) {
	path, err := paths.SettingsFile(env)
	if err != nil {
		return false, mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return false, mapCreateError(path, err)
	}

	info, statErr := os.Lstat(path)
	if statErr == nil {
		return false, refuseNonRegular(path, info)
	}

	if !os.IsNotExist(statErr) {
		return false, mapCreateError(path, statErr)
	}

	err = files.WriteAtomic(path, []byte(DefaultFileBody))
	if err != nil {
		return false, mapCreateError(path, err)
	}

	return true, nil
}

func refuseNonRegular(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return newError(KindWrite, path, files.ErrSymlink)
	}

	if !info.Mode().IsRegular() {
		return newError(KindWrite, path, files.ErrNotRegular)
	}

	return nil
}

func mapPathError(err error) error {
	if errors.Is(err, paths.ErrHomeUnset) {
		return newError(KindPath, "", err)
	}

	return fmt.Errorf("aws-cli-guard: settings: resolve config path: %w", err)
}

func mapCreateError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, err)
	}

	return newError(KindWrite, path, fmt.Errorf("aws-cli-guard: settings: %w", err))
}
