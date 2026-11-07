//go:build unix

package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Options configures aws-guard install.
type Options struct {
	Env        []string
	Prefix     string
	Executable string
}

// Run installs AWS CLI Guard into the chosen prefix.
func Run(opts Options, stdout io.Writer) error {
	inputs, err := prepareInstall(opts)
	if err != nil {
		return err
	}

	return writeInstallLayout(inputs, stdout)
}

func resolveExecutable(explicit string) (string, error) {
	if explicit != "" {
		resolved, err := filepath.EvalSymlinks(explicit)
		if err != nil {
			return "", fmt.Errorf("%w: %q", ErrResolveExecutable, explicit)
		}

		return resolved, nil
	}

	self, err := os.Executable()
	if err != nil {
		return "", ErrResolveExecutable
	}

	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return "", ErrResolveExecutable
	}

	return self, nil
}

func regularSibling(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", ErrSiblingMissing
	}

	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", ErrSiblingMissing
	}

	return resolved, nil
}

func refuseOccupied(paths []string) error {
	for _, path := range paths {
		occupied, err := pathOccupied(path)
		if err != nil {
			return err
		}

		if occupied {
			return ErrPathOccupied
		}
	}

	return nil
}
