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

	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatalf("ReadFile policy: %v", err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatalf("policy body mismatch")
	}

	data, err = os.ReadFile(filepath.Join(dir, "settings.toml"))
	if err != nil {
		t.Fatalf("ReadFile settings: %v", err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatalf("settings body = %q", string(data))
	}

	if !strings.Contains(stdout.String(), stayOffUpdatesStdout) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunInitAlreadyExistsWithStoredAuto(t *testing.T) {
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

	err = os.WriteFile(
		filepath.Join(dir, "settings.toml"),
		[]byte("version = 1\n\n[updates]\nauto = false\n"),
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	want := "aws-cli-guard: error: the policy file already exists"
	if strings.TrimSuffix(stderr.String(), "\n") != want {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunInitBothFilesPresentHeadlessSuccess(t *testing.T) {
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

	var stdout bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &stdout, &bytes.Buffer{})
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if !strings.Contains(stdout.String(), stayOffUpdatesStdout) {
		t.Fatalf("stdout = %q", stdout.String())
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

	var stdout bytes.Buffer

	status := cli.Run([]string{progName, initSubcommand}, env, &stdout, &bytes.Buffer{})
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if !strings.Contains(stdout.String(), "created settings file") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunInitUpdatesOnFromReader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	status := cli.Dispatch(
		[]string{progName, initSubcommand},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		panicIfConfirm,
		readLineReturns("on"),
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	data, err := os.ReadFile(filepath.Join(dir, "settings.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(data), "auto = true") {
		t.Fatalf("body = %q", data)
	}
}

func TestRunInitUpdatesMismatchWritesNothing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, initSubcommand},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
		readLineReturns("maybe"),
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	_, err := os.Stat(filepath.Join(dir, "config.toml"))
	if !os.IsNotExist(err) {
		t.Fatalf("policy should be missing: %v", err)
	}
}

const stayOffUpdatesStdout = "automatic updates stay off"

func readLineReturns(line string) func(string) (string, error) {
	return func(string) (string, error) {
		return line, nil
	}
}
