//go:build !awsguard_testpaths

package paths

// Production builds ignore AWS_CLI_GUARD_*_DIR; see ADR 0014.

func configDirOverride(_ []string) (string, bool) {
	return "", false
}

func stateDirOverride(_ []string) (string, bool) {
	return "", false
}

func cacheDirOverride(_ []string) (string, bool) {
	return "", false
}

func dirOverrideActive(_ []string) bool {
	return false
}
