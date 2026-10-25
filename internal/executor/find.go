package executor

import (
	"fmt"
	"os"
	"path/filepath"
)

// RealAWSEnv is the environment variable that overrides PATH discovery of the real AWS CLI.
const RealAWSEnv = "AWS_CLI_GUARD_REAL_AWS"

// FindRealAWS locates the real AWS CLI executable. It returns the path from RealAWSEnv when set,
// otherwise the first executable named aws on PATH that is not this wrapper.
func FindRealAWS() (string, error) {
	if override := os.Getenv(RealAWSEnv); override != "" {
		return override, nil
	}

	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: cannot resolve wrapper executable: %w", err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: cannot resolve wrapper executable: %w", err)
	}

	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, "aws")
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
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			resolved = candidate
		}
		if samePath(resolved, self) {
			continue
		}
		return candidate, nil
	}

	return "", fmt.Errorf("aws-cli-guard: real AWS CLI not found in PATH (set %s to override)", RealAWSEnv)
}

func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0o111 != 0
}

func samePath(a, b string) bool {
	stA, errA := os.Stat(a)
	stB, errB := os.Stat(b)
	if errA == nil && errB == nil {
		return os.SameFile(stA, stB)
	}
	cleanA, _ := filepath.Abs(a)
	cleanB, _ := filepath.Abs(b)
	return cleanA == cleanB
}
