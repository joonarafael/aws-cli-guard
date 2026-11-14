package settings_test

import (
	"aws-cli-guard/internal/settings"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUpdatesAbsentDefault(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, settings.DefaultFileBody)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Auto != settings.AutoAbsent {
		t.Fatalf("Auto = %v, want absent", got.Auto)
	}
}

func TestWriteAutoPreservesLogging(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[identity]
verify = false

[logging]
level = "verbose"
`)

	_, err := settings.WriteAuto(env, settings.AutoOn)
	if err != nil {
		t.Fatalf("WriteAuto: %v", err)
	}

	path := filepath.Join(settingsDirFromEnv(t, env), "settings.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, `level = "verbose"`) {
		t.Fatalf("body = %q", body)
	}

	if !strings.Contains(body, "auto = true") {
		t.Fatalf("body = %q", body)
	}
}

func TestEncodeBodyRejectsAbsent(t *testing.T) {
	t.Parallel()

	_, err := settings.EncodeBody(1, false, settings.LevelTight, settings.AutoAbsent)
	if err == nil {
		t.Fatal("expected error")
	}
}
