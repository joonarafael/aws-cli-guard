//go:build unix

package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Options configures aws-guard install and upgrade.
type Options struct {
	Env           []string
	Prefix        string
	Executable    string
	BeforeUpgrade func() error
}

// Run installs into the chosen prefix or upgrades when that prefix is already installed.
func Run(opts Options, stdout io.Writer) (Result, error) {
	return installOrUpgrade(opts, stdout)
}

// Upgrade replaces an existing installation in the chosen prefix.
func Upgrade(opts Options, stdout io.Writer) (Result, error) {
	return runUpgrade(opts, stdout)
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
