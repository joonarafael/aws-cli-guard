//go:build unix

package install

import (
	"aws-cli-guard/internal/environ"
	"debug/buildinfo"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverRealAWS returns the first executable aws on pathEnv that is not the install wrapper.
func DiscoverRealAWS(pathEnv, wrapperAWSPath string) (string, error) {
	wrapperAWSPath, err := absNoEval(wrapperAWSPath)
	if err != nil {
		return "", err
	}

	wrapperPkg := packagePath(wrapperAWSPath)

	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, nameAWS)

		info, err := os.Stat(candidate)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return "", fmt.Errorf("aws-cli-guard: stat %q: %w", candidate, err)
		}

		if info.IsDir() {
			continue
		}

		if info.Mode()&0o111 == 0 {
			continue
		}

		if isInstallWrapper(candidate, wrapperAWSPath, wrapperPkg) {
			continue
		}

		resolved, err := absNoEval(candidate)
		if err != nil {
			return "", err
		}

		return resolved, nil
	}

	return "", ErrRealAWSNotFound
}

func discoverPathEnv(env []string) string {
	value, ok := environ.Lookup("PATH", env)
	if !ok {
		return ""
	}

	return value
}

func packagePath(path string) string {
	info, err := buildinfo.ReadFile(path)
	if err != nil || info.Path == "" {
		return ""
	}

	return info.Path
}

func isInstallWrapper(candidate, wrapperAWSPath, wrapperPkg string) bool {
	if sameInstallPath(candidate, wrapperAWSPath) {
		return true
	}

	if wrapperPkg == "" {
		return false
	}

	info, err := buildinfo.ReadFile(candidate)
	if err != nil {
		return false
	}

	return info.Path == wrapperPkg
}

func sameInstallPath(leftPath, rightPath string) bool {
	resolvedLeft, err := filepath.EvalSymlinks(leftPath)
	if err != nil {
		resolvedLeft = leftPath
	}

	resolvedRight, err := filepath.EvalSymlinks(rightPath)
	if err != nil {
		resolvedRight = rightPath
	}

	stLeft, errLeft := os.Stat(resolvedLeft)

	stRight, errRight := os.Stat(resolvedRight)
	if errLeft == nil && errRight == nil {
		return os.SameFile(stLeft, stRight)
	}

	cleanLeft, err := filepath.Abs(resolvedLeft)
	if err != nil {
		cleanLeft = resolvedLeft
	}

	cleanRight, err := filepath.Abs(resolvedRight)
	if err != nil {
		cleanRight = resolvedRight
	}

	return cleanLeft == cleanRight
}

func absNoEval(path string) (string, error) {
	clean := filepath.Clean(path)

	resolved, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: resolve path %q: %w", path, err)
	}

	return resolved, nil
}
