package executor

import (
	"debug/buildinfo"
	"os"
	"path/filepath"
	"runtime/debug"
)

// RealAWSEnv is the environment variable that overrides PATH discovery of the real AWS CLI.
const RealAWSEnv = "AWS_CLI_GUARD_REAL_AWS"

// FindRealAWS locates the real AWS CLI executable. It returns the path from RealAWSEnv when set,
// otherwise the first executable named aws on PATH that is not an AWS CLI Guard wrapper.
func FindRealAWS() (string, error) {
	self, selfErr := resolvedSelf()

	ownPkg := ownPackagePath()

	if override := os.Getenv(RealAWSEnv); override != "" {
		if isWrapper(override, self, selfErr, ownPkg) {
			return "", ErrConfiguredWrapper
		}

		return override, nil
	}

	if selfErr != nil {
		return "", ErrResolveWrapper
	}

	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, "aws")

		//nolint:gosec // G703: PATH segments joined with fixed basename "aws".
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		if !isExecutable(info) {
			continue
		}

		if isWrapper(candidate, self, nil, ownPkg) {
			continue
		}

		return candidate, nil
	}

	return "", ErrRealAWSNotFound
}

func resolvedSelf() (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", ErrResolveWrapper
	}

	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return "", ErrResolveWrapper
	}

	return self, nil
}

func ownPackagePath() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Path == "" {
		return ""
	}

	return info.Path
}

func isWrapper(candidate, self string, selfErr error, ownPkg string) bool {
	if selfErr == nil && samePath(candidate, self) {
		return true
	}

	if ownPkg == "" {
		return false
	}

	info, err := buildinfo.ReadFile(candidate)
	if err != nil {
		return false
	}

	return info.Path == ownPkg
}

func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0o111 != 0
}

//nolint:gosec // G703: compares resolved PATH candidates for wrapper self-skip only.
func samePath(leftPath, rightPath string) bool {
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
