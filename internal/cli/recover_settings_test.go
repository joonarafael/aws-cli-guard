package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/terminal"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRecoverSettingsSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	settingsPath := filepath.Join(dir, "settings.toml")

	err := os.WriteFile(settingsPath, []byte("broken\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	confirm := func(phrase, _ string) error {
		if phrase != recoverSettings {
			return terminal.ErrMismatch
		}

		return nil
	}

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverSettings},
		env,
		&stdout,
		&bytes.Buffer{},
		confirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatal("settings not replaced")
	}

	if !strings.Contains(stdout.String(), "aside:") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRecoverSettingsMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	called := false
	confirm := func(string, string) error {
		called = true

		return nil
	}

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverSettings},
		env,
		&bytes.Buffer{},
		&stderr,
		confirm,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if called {
		t.Fatal("confirm should not run")
	}

	line := strings.TrimSuffix(stderr.String(), "\n")

	want := "aws-cli-guard: error: the settings file does not exist"
	if line != want {
		t.Fatalf("stderr = %q", line)
	}
}
