package settings_test

import (
	"aws-cli-guard/internal/settings"
	"errors"
	"strings"
	"testing"
)

func TestLoadLoggingLevelVerbose(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[logging]
level = "verbose"
`)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Level != settings.LevelVerbose {
		t.Fatalf("Level = %q, want verbose", got.Level)
	}
}

func TestLoadLoggingLevelTightExplicit(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[logging]
level = "tight"
`)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Level != settings.LevelTight {
		t.Fatalf("Level = %q, want tight", got.Level)
	}
}

func TestLoadLoggingTableMissingLevelInvalid(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[logging]
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Kind != settings.KindInvalid || setErr.Key != "logging.level" {
		t.Fatalf("got %v", err)
	}
}

func TestLoadLoggingLevelBadInvalid(t *testing.T) {
	t.Parallel()

	secret := "super-secret-log-level"
	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[logging]
level = "`+secret+`"
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if strings.Contains(err.Error(), secret) {
		t.Fatal("error echoed rejected level value")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Key != "logging.level" {
		t.Fatalf("got %v", err)
	}
}
