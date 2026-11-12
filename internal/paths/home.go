package paths

import (
	"aws-cli-guard/internal/environ"
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
)

var errEmptyUserHome = errors.New("aws-cli-guard: user home directory is empty")

const (
	homeOriginUserDatabase = "user database"
	homeOriginHOME         = "HOME"
)

type homeLookup func() (string, error)

type homeResolution struct {
	dir    string
	origin string
	ok     bool
}

func defaultUserHomeLookup() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: user lookup: %w", err)
	}

	if current.HomeDir == "" {
		return "", errEmptyUserHome
	}

	return current.HomeDir, nil
}

func resolveHome(env []string, lookup homeLookup) homeResolution {
	if dirOverrideActive(env) {
		return homeResolution{} //nolint:exhaustruct_v5 // not resolved under test overrides.
	}

	fromUser, err := lookup()
	if err == nil && isAbsoluteHome(fromUser) {
		return homeResolution{dir: fromUser, origin: homeOriginUserDatabase, ok: true}
	}

	fallback, ok := environ.Lookup("HOME", env)
	if ok && isAbsoluteHome(fallback) {
		return homeResolution{dir: fallback, origin: homeOriginHOME, ok: true}
	}

	return homeResolution{} //nolint:exhaustruct_v5 // not resolved.
}

func homeDir(env []string, lookup homeLookup) (string, error) {
	res := resolveHome(env, lookup)
	if !res.ok {
		return "", ErrHomeUnset
	}

	return res.dir, nil
}

// HomeOrigin reports whether guard-owned paths used the user database or HOME.
// The second result is false when home cannot be resolved or a test-path override is set.
func HomeOrigin(env []string) (string, bool) {
	res := resolveHome(env, defaultUserHomeLookup)
	if !res.ok {
		return "", false
	}

	return res.origin, true
}

func isAbsoluteHome(home string) bool {
	return home != "" && filepath.IsAbs(home)
}

func configDir(env []string) (string, error) {
	return configDirWithLookup(env, defaultUserHomeLookup)
}

func configDirWithLookup(env []string, lookup homeLookup) (string, error) {
	if override, ok := configDirOverride(env); ok {
		return override, nil
	}

	home, err := homeDir(env, lookup)
	if err != nil {
		return "", fmt.Errorf("%w for config", err)
	}

	return filepath.Join(home, ".config", "aws-cli-guard"), nil
}

func stateDir(env []string) (string, error) {
	return stateDirWithLookup(env, defaultUserHomeLookup)
}

func stateDirWithLookup(env []string, lookup homeLookup) (string, error) {
	if override, ok := stateDirOverride(env); ok {
		return override, nil
	}

	home, err := homeDir(env, lookup)
	if err != nil {
		return "", fmt.Errorf("%w for state", err)
	}

	return filepath.Join(home, ".local", "state", "aws-cli-guard"), nil
}

func cacheRoot(env []string) (string, error) {
	return cacheRootWithLookup(env, defaultUserHomeLookup)
}

func cacheRootWithLookup(env []string, lookup homeLookup) (string, error) {
	if override, ok := cacheDirOverride(env); ok {
		return override, nil
	}

	home, err := homeDir(env, lookup)
	if err != nil {
		return "", fmt.Errorf("%w for cache", err)
	}

	return filepath.Join(home, ".cache", "aws-cli-guard"), nil
}
