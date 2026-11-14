package update

import (
	"aws-cli-guard/internal/files"
	"path/filepath"
)

//nolint:wrapcheck // mkdir failure skips the helper quietly.
func ensureUpdatesCacheDir(updatesPath string) error {
	if updatesPath == "" {
		return nil
	}

	return files.MkdirPrivate(filepath.Dir(updatesPath))
}
