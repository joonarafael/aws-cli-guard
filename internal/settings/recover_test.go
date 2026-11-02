package settings_test

import (
	"aws-cli-guard/internal/settings"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecoverReplacesGarbage(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, "not-valid = [[[\n"+fixtureSecret+"\n")

	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	aside, err := settings.Recover(env, now)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(settingsDirFromEnv(t, env), "settings.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatalf("settings body mismatch")
	}

	asideData, err := os.ReadFile(aside)
	if err != nil {
		t.Fatalf("ReadFile aside: %v", err)
	}

	if string(asideData) != "not-valid = [[[\n"+fixtureSecret+"\n" {
		t.Fatalf("aside should preserve old bytes")
	}
}

func TestRecoverMissingFile(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := settingsDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	_, err = settings.Recover(env, time.Now())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, settings.ErrMissing) {
		t.Fatalf("got %v", err)
	}
}
