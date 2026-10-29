package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/terminal"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	recoverSubcommand = "recover"
	recoverPolicy     = "policy"
	recoverGrants     = "grants"
	recoverCache      = "cache"
)

func TestRunRecoverPolicySuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	policyPath := filepath.Join(dir, "config.toml")

	err := os.WriteFile(policyPath, []byte("broken\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	confirm := func(phrase, _ string) error {
		if phrase != recoverPolicy {
			return terminal.ErrMismatch
		}

		return nil
	}

	var stdout, stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverPolicy},
		env,
		&stdout,
		&stderr,
		confirm,
	)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}

	data, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatal("policy not replaced")
	}

	if !strings.Contains(stdout.String(), "aside:") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRecoverPolicyMissing(t *testing.T) {
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
		[]string{progName, recoverSubcommand, recoverPolicy},
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

	want := "aws-cli-guard: error: the policy file does not exist"
	if line != want {
		t.Fatalf("stderr = %q", line)
	}
}

func TestRunRecoverMismatchLeavesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	body := []byte("keep-me\n")
	policyPath := filepath.Join(dir, "config.toml")

	err := os.WriteFile(policyPath, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	confirm := func(string, string) error {
		return terminal.ErrMismatch
	}

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverPolicy},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		confirm,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	got, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, body) {
		t.Fatal("file changed on mismatch")
	}
}

func TestRunRecoverGrantsSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_STATE_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	grantsPath := filepath.Join(dir, "grants.json")

	err = os.WriteFile(grantsPath, []byte("{"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	confirm := func(phrase, _ string) error {
		if phrase != recoverGrants {
			return terminal.ErrMismatch
		}

		return nil
	}

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverGrants},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		confirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(file.Grants) != 0 {
		t.Fatalf("grants = %+v", file.Grants)
	}
}

func TestRunRecoverCacheAbsentOK(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CACHE_DIR=" + filepath.Join(base, "cache")}

	called := false
	confirm := func(string, string) error {
		called = true

		return nil
	}

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverCache},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		confirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if called {
		t.Fatal("confirm should not run for absent cache")
	}
}

func TestRunRecoverUsageNoEcho(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, "settings"},
		nil,
		&bytes.Buffer{},
		&stderr,
		func(string, string) error { return nil },
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if strings.Contains(stderr.String(), "settings") {
		t.Fatalf("stderr must not echo argv: %q", stderr.String())
	}
}

func TestRunRecoverExtraArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, recoverSubcommand, recoverPolicy, "extra"},
		nil,
		&bytes.Buffer{},
		&stderr,
		func(string, string) error { return nil },
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if strings.Contains(stderr.String(), "extra") {
		t.Fatalf("stderr must not echo argv")
	}
}
