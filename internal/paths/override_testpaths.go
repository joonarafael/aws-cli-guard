//go:build awsguard_testpaths

package paths

import "aws-cli-guard/internal/environ"

const (
	configDirEnv = "AWS_CLI_GUARD_CONFIG_DIR"
	stateDirEnv  = "AWS_CLI_GUARD_STATE_DIR"
	cacheDirEnv  = "AWS_CLI_GUARD_CACHE_DIR"
)

func configDirOverride(env []string) (string, bool) {
	return dirOverride(configDirEnv, env)
}

func stateDirOverride(env []string) (string, bool) {
	return dirOverride(stateDirEnv, env)
}

func cacheDirOverride(env []string) (string, bool) {
	return dirOverride(cacheDirEnv, env)
}

func dirOverride(key string, env []string) (string, bool) {
	override, ok := environ.Lookup(key, env)
	if !ok || override == "" {
		return "", false
	}

	return override, true
}

func dirOverrideActive(env []string) bool {
	_, present := configDirOverride(env)
	if present {
		return true
	}

	_, present = stateDirOverride(env)
	if present {
		return true
	}

	_, present = cacheDirOverride(env)

	return present
}
