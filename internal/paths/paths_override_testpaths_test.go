//go:build awsguard_testpaths

//nolint:testpackage // exercises test-only directory overrides.
package paths

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirOverrideReplacesHome(t *testing.T) {
	t.Parallel()

	env := []string{
		"AWS_CLI_GUARD_CONFIG_DIR=/custom/config",
		ignoredHomeEnvEntry,
	}

	dir, err := configDirWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("configDirWithLookup: %v", err)
	}

	want := "/custom/config"
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}

func TestConfigDirEmptyOverrideFallsThrough(t *testing.T) {
	t.Parallel()

	env := []string{
		"AWS_CLI_GUARD_CONFIG_DIR=",
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
}

func TestStateDirOverride(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_STATE_DIR=/state/override"}

	dir, err := stateDirWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("stateDirWithLookup: %v", err)
	}

	if dir != "/state/override" {
		t.Fatalf("got %q", dir)
	}
}

func TestCacheDirOverride(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_CACHE_DIR=/cache/override"}

	dir, err := cacheRootWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("cacheRootWithLookup: %v", err)
	}

	if dir != "/cache/override" {
		t.Fatalf("got %q", dir)
	}
}

func TestRelativeOverrideIsAbsoluteInConfigFile(t *testing.T) {
	t.Parallel()

	env := []string{
		"AWS_CLI_GUARD_CONFIG_DIR=rel/config",
		"HOME=/home/user",
	}

	got, err := ConfigFile(env)
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute path, got %q", got)
	}

	if !strings.HasSuffix(got, filepath.FromSlash("rel/config/config.toml")) {
		t.Fatalf("unexpected path %q", got)
	}
}

func TestXDGIgnoredWithOverride(t *testing.T) {
	t.Parallel()

	env := []string{
		"XDG_CONFIG_HOME=/xdg/config",
		"HOME=/home/user",
	}

	dir, err := configDirWithLookup(env, lookupReturning("/db/home"))
	if err != nil {
		t.Fatalf("configDirWithLookup: %v", err)
	}

	want := filepath.Join("/db/home", ".config", "aws-cli-guard")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}
