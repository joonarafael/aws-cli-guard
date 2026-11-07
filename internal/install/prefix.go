//go:build unix

package install

import (
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolvePrefix(env []string, userPrefix string) (string, error) {
	if userPrefix == "" {
		dir, err := paths.DefaultInstallPrefix(env)
		if err != nil {
			if errors.Is(err, paths.ErrHomeUnset) {
				return "", ErrHomeUnset
			}

			return "", fmt.Errorf("aws-cli-guard: default install prefix: %w", err)
		}

		return dir, nil
	}

	resolved, err := filepath.Abs(userPrefix)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: resolve install prefix %q: %w", userPrefix, err)
	}

	return resolved, nil
}

func validatePrefixChars(prefix string) error {
	if strings.ContainsAny(prefix, "\r\n'") {
		return ErrPrefixInvalid
	}

	return nil
}

const installDirMode = 0o755

type installLayout struct {
	aws           string
	awsGuard      string
	disableScript string
	aside         string
	marker        string
}

func layoutPaths(prefix string) installLayout {
	return installLayout{
		aws:           filepath.Join(prefix, nameAWS),
		awsGuard:      filepath.Join(prefix, nameAWSGuard),
		disableScript: filepath.Join(prefix, nameDisableScript),
		aside:         filepath.Join(prefix, nameAside),
		marker:        filepath.Join(prefix, nameDisabledMarker),
	}
}

func reservedInstallPaths(prefix string) []string {
	layout := layoutPaths(prefix)

	return []string{
		layout.aws,
		layout.awsGuard,
		layout.disableScript,
		filepath.Join(prefix, nameManifest),
		layout.aside,
		layout.marker,
	}
}

func pathOccupied(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("aws-cli-guard: stat %q: %w", path, err)
	}

	return true, nil
}

func realAWSConflictsPrefix(realAWS string, reserved []string) (bool, error) {
	cleanReal, err := absNoEval(realAWS)
	if err != nil {
		return false, err
	}

	for _, reservedPath := range reserved {
		cleanReserved, err := absNoEval(reservedPath)
		if err != nil {
			return false, err
		}

		if cleanReal == cleanReserved {
			return true, nil
		}

		stReal, errReal := os.Stat(cleanReal)

		stReserved, errReserved := os.Stat(cleanReserved)
		if errReal == nil && errReserved == nil {
			if os.SameFile(stReal, stReserved) {
				return true, nil
			}
		}
	}

	return false, nil
}

func ensurePrefixDirectory(prefix string) error {
	info, err := os.Lstat(prefix)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(prefix, installDirMode)
			if err != nil {
				return fmt.Errorf("aws-cli-guard: create install prefix %q: %w", prefix, err)
			}

			return nil
		}

		return fmt.Errorf("aws-cli-guard: stat install prefix %q: %w", prefix, err)
	}

	if !info.IsDir() {
		return ErrPrefixNotDirectory
	}

	return nil
}
