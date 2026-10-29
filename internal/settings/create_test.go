package settings_test

import (
	"aws-cli-guard/internal/settings"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWritesDefaultSettings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	created, err := settings.Create(env)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !created {
		t.Fatal("expected created true")
	}

	path := filepath.Join(dir, "settings.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatalf("body mismatch")
	}
}

func TestCreateLeavesExistingSettings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "settings.toml")

	err = os.WriteFile(path, []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	created, err := settings.Create(env)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if created {
		t.Fatal("expected created false")
	}
}
