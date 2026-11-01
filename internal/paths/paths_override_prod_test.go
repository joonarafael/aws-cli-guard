//go:build !awsguard_testpaths

//nolint:testpackage // verifies production builds ignore directory override variables.
package paths

import (
	"path/filepath"
	"testing"
)

func TestConfigDirIgnoresGuardDirEnvInProductionBuild(t *testing.T) {
	t.Parallel()

	env := []string{
		"AWS_CLI_GUARD_CONFIG_DIR=/forged/config",
		"AWS_CLI_GUARD_STATE_DIR=/forged/state",
		"AWS_CLI_GUARD_CACHE_DIR=/forged/cache",
		"HOME=/env/home",
	}

	dir, err := configDirWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("configDirWithLookup: %v", err)
	}

	want := filepath.Join("/db/home", ".config", "aws-cli-guard")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}

	state, err := stateDirWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("stateDirWithLookup: %v", err)
	}

	wantState := filepath.Join("/db/home", ".local", "state", "aws-cli-guard")
	if state != wantState {
		t.Fatalf("state got %q, want %q", state, wantState)
	}

	cache, err := cacheRootWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("cacheRootWithLookup: %v", err)
	}

	wantCache := filepath.Join("/db/home", ".cache", "aws-cli-guard")
	if cache != wantCache {
		t.Fatalf("cache got %q, want %q", cache, wantCache)
	}
}
