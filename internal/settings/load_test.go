package settings_test

import (
	"aws-cli-guard/internal/settings"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultFileBody(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, settings.DefaultFileBody)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Verify {
		t.Fatal("Verify = true, want false")
	}

	if got.Version != 1 {
		t.Fatalf("Version = %d", got.Version)
	}
}

func TestLoadVersionOnly(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, "version = 1\n")

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Verify {
		t.Fatal("Verify = true, want false")
	}
}

func TestLoadMissingFileDefaults(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := settingsDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Verify || got.Version != 1 {
		t.Fatalf("got %+v", got)
	}

	path := filepath.Join(dir, "settings.toml")

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatalf("Stat: %v", err)
	}
}

func TestLoadLoosePermissionsStillLoads(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettingsPerm(t, env, settings.DefaultFileBody, 0o644)

	_, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
}
