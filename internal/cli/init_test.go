package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const initSubcommand = "init"

func TestRunInitCreatesFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	var stdout, stderr bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	policyPath := filepath.Join(dir, "config.toml")

	data, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("ReadFile policy: %v", err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatalf("policy body mismatch")
	}

	settingsPath := filepath.Join(dir, "settings.toml")

	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile settings: %v", err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatalf("settings body mismatch")
	}

	if !strings.Contains(stdout.String(), "created policy file") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunInitAlreadyExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "settings.toml"), []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &stdout, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	if stdout.Len() != 0 {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}

	line := strings.TrimSuffix(stderr.String(), "\n")

	want := "aws-cli-guard: error: the policy file already exists"
	if line != want {
		t.Fatalf("stderr = %q want %q", line, want)
	}
}

func TestRunInitSettingsOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}

	if !strings.Contains(stdout.String(), "policy file left unchanged") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	if !strings.Contains(stdout.String(), "created settings file") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunInitExtraArgsUsage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand, "extra"}, env, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	assertUsageLine(t, stderr.String())

	if strings.Contains(stderr.String(), "extra") {
		t.Fatalf("stderr must not echo argv: %q", stderr.String())
	}
}
