//go:build unix

package cli_test

import (
	"aws-cli-guard/internal/cli"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

const (
	recoverChildEnv = "AWS_CLI_GUARD_RECOVER_CHILD"
	exitUnchanged   = 11
)

func TestMain(m *testing.M) {
	if os.Getenv(recoverChildEnv) == "1" {
		dir := os.Getenv("AWS_CLI_GUARD_CONFIG_DIR")
		policyPath := filepath.Join(dir, "config.toml")

		before, err := os.ReadFile(policyPath)
		if err != nil {
			os.Exit(12)
		}

		status := cli.Run(
			[]string{"aws-guard", "recover", recoverPolicy},
			os.Environ(),
			os.Stdout,
			os.Stderr,
		)

		after, err := os.ReadFile(policyPath)
		if err != nil {
			os.Exit(13)
		}

		if status == 0 || !bytes.Equal(before, after) {
			os.Exit(exitUnchanged + 1)
		}

		os.Exit(exitUnchanged)
	}

	os.Exit(m.Run())
}

//nolint:paralleltest // Setsid changes process session.
func TestRecoverPolicyHeadlessLeavesFile(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "config.toml")
	body := []byte("version = 1\n")

	err := os.WriteFile(policyPath, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	exitCode := runHeadlessRecoverChild(t, dir)
	if exitCode != exitUnchanged {
		t.Fatalf("exit = %d, want %d", exitCode, exitUnchanged)
	}

	got, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, body) {
		t.Fatal("policy file must be unchanged")
	}
}

func runHeadlessRecoverChild(t *testing.T, configDir string) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^$")

	cmd.Env = append(os.Environ(),
		recoverChildEnv+"=1",
		"AWS_CLI_GUARD_CONFIG_DIR="+configDir,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	//nolint:exhaustruct_v5 // only Setsid is required for this test.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	_, err = stdin.Write([]byte(recoverPolicy + "\n"))
	if err != nil {
		t.Fatal(err)
	}

	_ = stdin.Close()

	err = cmd.Wait()

	if ctx.Err() != nil {
		t.Fatal("child did not exit")
	}

	exitErr := &exec.ExitError{}

	ok := errors.As(err, &exitErr)
	if !ok {
		t.Fatalf("wait: %v", err)
	}

	return exitErr.ExitCode()
}
