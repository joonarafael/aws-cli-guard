package policy

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultFileBody is the init and recover default for config.toml (schema version 1).
const DefaultFileBody = `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"
`

// Create writes the default policy file when it is missing. It returns created true when
// a new file was written. An existing regular file is left unchanged (created false, nil error).
func Create(env []string) (bool, error) {
	path, err := paths.ConfigFile(env)
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
		return newError(KindWrite, path, "", 0, 0, files.ErrSymlink)
	}

	if !info.Mode().IsRegular() {
		return newError(KindWrite, path, "", 0, 0, files.ErrNotRegular)
	}

	return nil
}

func mapCreateError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: policy: %w", err))
}
