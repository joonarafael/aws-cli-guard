package settings_test

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/settings"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVerifyTrueInvalid(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[identity]
verify = true
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Kind != settings.KindInvalid || setErr.Key != "identity.verify" {
		t.Fatalf("got %v", err)
	}

	assertNoSecretInError(t, err)
}

func TestLoadIdentityWithoutVerifyInvalid(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[identity]
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Key != "identity.verify" {
		t.Fatalf("got %v", err)
	}
}

func TestLoadUpdatesAutoFalse(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[updates]
auto = false
`)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Auto != settings.AutoOff {
		t.Fatalf("Auto = %v, want off", got.Auto)
	}
}

func TestLoadUpdatesAutoTrue(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[updates]
auto = true
`)

	got, err := settings.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Auto != settings.AutoOn {
		t.Fatalf("Auto = %v, want on", got.Auto)
	}
}

func TestLoadUpdatesWithoutAutoInvalid(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[updates]
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Kind != settings.KindInvalid || setErr.Key != "updates.auto" {
		t.Fatalf("got %v", err)
	}
}

func TestLoadUnsupportedVersion(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, "version = 2\n")

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadSymlinkUnreadable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := settingsDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "real.toml")

	err = os.WriteFile(target, []byte(settings.DefaultFileBody), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "settings.toml")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatal(err)
	}

	_, err = settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	setErr, ok := errors.AsType[*settings.Error](err)
	if !ok || setErr.Kind != settings.KindRead {
		t.Fatalf("got %v", err)
	}

	if !errors.Is(err, files.ErrSymlink) {
		t.Fatalf("errors.Is symlink: got %v", err)
	}
}

func TestLoadTooLarge(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := settingsDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	body := make([]byte, (1<<20)+1)
	copy(body, "version = 1\n")

	path := filepath.Join(dir, "settings.toml")

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, settings.ErrTooLarge) {
		t.Fatalf("got %v", err)
	}
}

func TestLoadBadSyntaxSecretSafe(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, "not-valid = [[[\n"+fixtureSecret+"\n")

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertNoSecretInError(t, err)
}
