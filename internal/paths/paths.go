package paths

import (
	"aws-cli-guard/internal/environ"
	"fmt"
	"path/filepath"
)

const (
	configDirEnv = "AWS_CLI_GUARD_CONFIG_DIR"
	stateDirEnv  = "AWS_CLI_GUARD_STATE_DIR"
	cacheDirEnv  = "AWS_CLI_GUARD_CACHE_DIR"

	xdgConfigHome = "XDG_CONFIG_HOME"
	xdgStateHome  = "XDG_STATE_HOME"
	xdgCacheHome  = "XDG_CACHE_HOME"
)

// ConfigFile returns the absolute path to the user policy file.
func ConfigFile(env []string) (string, error) {
	dir, err := configDir(env)
	if err != nil {
		return "", err
	}

	return abs(filepath.Join(dir, "config.toml"))
}

// GrantsFile returns the absolute path to the grants file.
func GrantsFile(env []string) (string, error) {
	dir, err := stateDir(env)
	if err != nil {
		return "", err
	}

	return abs(filepath.Join(dir, "grants.json"))
}

// CacheDir returns the absolute path to the cache directory.
func CacheDir(env []string) (string, error) {
	dir, err := cacheRoot(env)
	if err != nil {
		return "", err
	}

	return abs(dir)
}

func configDir(env []string) (string, error) {
	if override, ok := environ.Lookup(configDirEnv, env); ok && override != "" {
		return override, nil
	}

	if xdg, ok := environ.Lookup(xdgConfigHome, env); ok && xdg != "" {
		return filepath.Join(xdg, "aws-cli-guard"), nil
	}

	home, ok := environ.Lookup("HOME", env)
	if !ok || home == "" {
		return "", fmt.Errorf("%w for config", ErrHomeUnset)
	}

	return filepath.Join(home, ".config", "aws-cli-guard"), nil
}

func stateDir(env []string) (string, error) {
	if override, ok := environ.Lookup(stateDirEnv, env); ok && override != "" {
		return override, nil
	}

	if xdg, ok := environ.Lookup(xdgStateHome, env); ok && xdg != "" {
		return filepath.Join(xdg, "aws-cli-guard"), nil
	}

	home, ok := environ.Lookup("HOME", env)
	if !ok || home == "" {
		return "", fmt.Errorf("%w for state", ErrHomeUnset)
	}

	return filepath.Join(home, ".local", "state", "aws-cli-guard"), nil
}

func cacheRoot(env []string) (string, error) {
	if override, ok := environ.Lookup(cacheDirEnv, env); ok && override != "" {
		return override, nil
	}

	if xdg, ok := environ.Lookup(xdgCacheHome, env); ok && xdg != "" {
		return filepath.Join(xdg, "aws-cli-guard"), nil
	}

	home, ok := environ.Lookup("HOME", env)
	if !ok || home == "" {
		return "", fmt.Errorf("%w for cache", ErrHomeUnset)
	}

	return filepath.Join(home, ".cache", "aws-cli-guard"), nil
}

func abs(path string) (string, error) {
	clean := filepath.Clean(path)

	resolved, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: resolve path %q: %w", path, err)
	}

	return resolved, nil
}
