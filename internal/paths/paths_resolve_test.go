//nolint:testpackage // exercises unexported path resolution helpers with injected lookups.
package paths

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

var errLookupFailed = errors.New("lookup failed")

const ignoredHomeEnvEntry = "HOME=/should/not/use"

func lookupReturning(home string) homeLookup {
	return func() (string, error) {
		return home, nil
	}
}

func lookupFailing() homeLookup {
	return func() (string, error) {
		return "", errLookupFailed
	}
}

func TestConfigDirUserHomeWinsOverEnv(t *testing.T) {
	t.Parallel()

	env := []string{
		"HOME=/env/home",
		"XDG_CONFIG_HOME=/xdg/config",
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

func TestHomeDirFallbackWhenLookupFails(t *testing.T) {
	t.Parallel()

	env := []string{"HOME=/fallback/home"}

	home, err := homeDir(env, lookupFailing())
	if err != nil {
		t.Fatalf("homeDir: %v", err)
	}

	if home != "/fallback/home" {
		t.Fatalf("got %q, want /fallback/home", home)
	}
}

func TestHomeDirFallbackWhenUserHomeRelative(t *testing.T) {
	t.Parallel()

	env := []string{"HOME=/fallback/home"}

	home, err := homeDir(env, lookupReturning("relative/home"))
	if err != nil {
		t.Fatalf("homeDir: %v", err)
	}

	if home != "/fallback/home" {
		t.Fatalf("got %q, want /fallback/home", home)
	}
}

func TestHomeDirErrWhenNoAbsoluteHome(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		env  []string
	}{
		{name: "empty env", env: nil},
		{name: "empty home", env: []string{"HOME="}},
		{name: "relative home", env: []string{"HOME=rel/home"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := homeDir(testCase.env, lookupFailing())
			if err == nil {
				t.Fatal("expected error")
			}

			if !errors.Is(err, ErrHomeUnset) {
				t.Fatalf("expected ErrHomeUnset, got %v", err)
			}
		})
	}
}

func TestGuardPathsLayout(t *testing.T) {
	t.Parallel()

	env := []string{ignoredHomeEnvEntry}
	lookup := lookupReturning("/db/home")

	configDir, err := configDirWithLookup(env, lookup)
	if err != nil {
		t.Fatalf("configDirWithLookup: %v", err)
	}

	stateDir, err := stateDirWithLookup(env, lookup)
	if err != nil {
		t.Fatalf("stateDirWithLookup: %v", err)
	}

	cacheDir, err := cacheRootWithLookup(env, lookup)
	if err != nil {
		t.Fatalf("cacheRootWithLookup: %v", err)
	}

	configPath, err := abs(filepath.Join(configDir, "config.toml"))
	if err != nil {
		t.Fatalf("abs config: %v", err)
	}

	settingsViaLookup, err := abs(filepath.Join(configDir, "settings.toml"))
	if err != nil {
		t.Fatalf("abs settings: %v", err)
	}

	grantsViaLookup, err := abs(filepath.Join(stateDir, "grants.json"))
	if err != nil {
		t.Fatalf("abs grants: %v", err)
	}

	cacheViaLookup, err := abs(cacheDir)
	if err != nil {
		t.Fatalf("abs cache: %v", err)
	}

	wantConfig := filepath.Join("/db/home", ".config", "aws-cli-guard", "config.toml")
	wantSettings := filepath.Join("/db/home", ".config", "aws-cli-guard", "settings.toml")
	wantGrants := filepath.Join("/db/home", ".local", "state", "aws-cli-guard", "grants.json")
	wantCache := filepath.Join("/db/home", ".cache", "aws-cli-guard")

	if configPath != wantConfig {
		t.Fatalf("config got %q, want %q", configPath, wantConfig)
	}

	if settingsViaLookup != wantSettings {
		t.Fatalf("settings got %q, want %q", settingsViaLookup, wantSettings)
	}

	if grantsViaLookup != wantGrants {
		t.Fatalf("grants got %q, want %q", grantsViaLookup, wantGrants)
	}

	if cacheViaLookup != wantCache {
		t.Fatalf("cache got %q, want %q", cacheViaLookup, wantCache)
	}
}

func TestSessionsFileLayout(t *testing.T) {
	t.Parallel()

	env := []string{ignoredHomeEnvEntry}

	got, err := SessionsFile(env)
	if err != nil {
		t.Fatalf("SessionsFile: %v", err)
	}

	want := filepath.Join("/db/home", ".local", "state", "aws-cli-guard", "sessions.json")
	if got != want {
		t.Fatalf("sessions got %q, want %q", got, want)
	}
}

func TestDefaultInstallPrefixLayout(t *testing.T) {
	t.Parallel()

	env := []string{ignoredHomeEnvEntry}
	lookup := lookupReturning("/db/home")

	installViaLookup, err := defaultInstallPrefixWithLookup(env, lookup)
	if err != nil {
		t.Fatalf("defaultInstallPrefixWithLookup: %v", err)
	}

	wantInstall := filepath.Join("/db/home", ".local", "share", "aws-cli-guard")
	if installViaLookup != wantInstall {
		t.Fatalf("install prefix got %q, want %q", installViaLookup, wantInstall)
	}
}

func TestDefaultInstallPrefixErrHomeUnset(t *testing.T) {
	t.Parallel()

	_, err := defaultInstallPrefixWithLookup(nil, lookupFailing())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrHomeUnset) {
		t.Fatalf("expected ErrHomeUnset, got %v", err)
	}
}

func TestConfigFileErrHomeUnset(t *testing.T) {
	t.Parallel()

	// Force failure through public API only when both lookup and HOME fail.
	_, err := configDirWithLookup(nil, lookupFailing())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrHomeUnset) {
		t.Fatalf("expected ErrHomeUnset, got %v", err)
	}

	if !strings.Contains(err.Error(), "for config") {
		t.Fatalf("expected config scope in error, got %v", err)
	}
}

func TestConfigFileRealLookupSmoke(t *testing.T) {
	t.Parallel()

	got, err := ConfigFile(nil)
	if err != nil {
		if !errors.Is(err, ErrHomeUnset) {
			t.Fatalf("ConfigFile: %v", err)
		}

		return
	}

	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute path, got %q", got)
	}
}
