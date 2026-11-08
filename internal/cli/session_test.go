package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/sessions"
	"aws-cli-guard/internal/terminal"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunOffHeadlessWritesNothing(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_STATE_DIR=" + t.TempDir()}

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, "off"},
		env,
		&bytes.Buffer{},
		&stderr,
		rejectConfirm,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	_, err := os.Stat(filepath.Join(envStateFrom(env), "sessions.json"))
	if err == nil {
		t.Fatal("sessions.json must not be created")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Stat: %v", err)
	}
}

func TestRunOffExtraArgWritesNothing(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_STATE_DIR=" + t.TempDir()}

	status := cli.Dispatch(
		[]string{progName, "off", cliExtraArg},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		nilConfirm,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	_, err := os.Stat(filepath.Join(envStateFrom(env), "sessions.json"))
	if !os.IsNotExist(err) && err != nil {
		t.Fatalf("Stat: %v", err)
	}
}

func TestRunOnNoTerminalAppends(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_STATE_DIR=" + t.TempDir()}

	status := cli.Dispatch(
		[]string{progName, "on"},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		nilConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	got, err := sessions.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Records) != 1 || got.Records[0].Effect != sessions.EffectOn {
		t.Fatalf("records: %+v", got.Records)
	}
}

func envStateFrom(env []string) string {
	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_STATE_DIR="
		if len(entry) > len(prefix) && entry[:len(prefix)] == prefix {
			return entry[len(prefix):]
		}
	}

	return ""
}

func nilConfirm(string, string) error {
	return nil
}

func rejectConfirm(string, string) error {
	return terminal.ErrMismatch
}
