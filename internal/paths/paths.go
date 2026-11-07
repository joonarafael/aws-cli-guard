package paths

import (
	"fmt"
	"path/filepath"
)

// ConfigFile returns the absolute path to the user policy file.
func ConfigFile(env []string) (string, error) {
	dir, err := configDir(env)
	if err != nil {
		return "", err
	}

	return abs(filepath.Join(dir, "config.toml"))
}

// SettingsFile returns the absolute path to the preferences file.
func SettingsFile(env []string) (string, error) {
	dir, err := configDir(env)
	if err != nil {
		return "", err
	}

	return abs(filepath.Join(dir, "settings.toml"))
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

// DefaultInstallPrefix returns the default install directory for aws-guard install.
func DefaultInstallPrefix(env []string) (string, error) {
	return defaultInstallPrefixWithLookup(env, defaultUserHomeLookup)
}

func defaultInstallPrefixWithLookup(env []string, lookup homeLookup) (string, error) {
	home, err := homeDir(env, lookup)
	if err != nil {
		return "", err
	}

	return abs(filepath.Join(home, ".local", "share", "aws-cli-guard"))
}

func abs(path string) (string, error) {
	clean := filepath.Clean(path)

	resolved, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: resolve path %q: %w", path, err)
	}

	return resolved, nil
}
