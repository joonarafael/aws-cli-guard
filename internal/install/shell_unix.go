//go:build unix

package install

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/output"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	shellStartupMaxSize = 1 << 20
	shellStartupMode    = 0o644
)

// ApplyShell edits the shell startup file for Bash or zsh.
func ApplyShell(opts ShellOptions, stdout io.Writer) (ShellResult, error) {
	prefix, err := resolvePrefix(opts.Env, opts.Prefix)
	if err != nil {
		return ShellResult{}, err
	}

	err = validatePrefixChars(prefix)
	if err != nil {
		return ShellResult{}, err
	}

	err = PrefixUpgradeable(prefix)
	if err != nil {
		return ShellResult{}, err
	}

	startupPath, err := shellStartupPath(opts.Home, opts.Shell)
	if err != nil {
		return ShellResult{}, err
	}

	content, mode, missing, err := readShellStartup(startupPath)
	if err != nil {
		return ShellResult{}, err
	}

	if opts.Remove && missing {
		output.Detail(stdout, "shell startup left unchanged")

		return ShellResult{Changed: false, Path: startupPath}, nil
	}

	edited, changed, err := editShellStartup(content, prefix, opts.Remove)
	if err != nil {
		return ShellResult{}, err
	}

	if !changed {
		output.Detail(stdout, "shell startup left unchanged")

		return ShellResult{Changed: false, Path: startupPath}, nil
	}

	if missing {
		mode = shellStartupMode
	}

	err = files.WriteAtomicMode(startupPath, []byte(edited), mode)
	if err != nil {
		return ShellResult{}, fmt.Errorf("%w: %w", ErrShellStartupUpdate, err)
	}

	output.Detail(stdout, "updated shell startup")
	output.Detail(stdout, "startup: "+startupPath)

	return ShellResult{Changed: true, Path: startupPath}, nil
}

func shellStartupPath(home string, shell ShellName) (string, error) {
	switch shell {
	case ShellBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellZsh:
		return filepath.Join(home, ".zshrc"), nil
	default:
		return "", ErrShellNameUnsupported
	}
}

func readShellStartup(path string) (string, os.FileMode, bool, error) {
	data, fileMode, readErr := files.ReadRegular(path)
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return "", shellStartupMode, true, nil
		}

		if errors.Is(readErr, files.ErrSymlink) || errors.Is(readErr, files.ErrNotRegular) {
			return "", 0, false, fmt.Errorf("aws-cli-guard: shell startup: %w", readErr)
		}

		return "", 0, false, fmt.Errorf("%w: %w", ErrShellStartupUpdate, readErr)
	}

	if len(data) > shellStartupMaxSize {
		return "", 0, false, ErrShellStartupTooLarge
	}

	return string(data), fileMode, false, nil
}
