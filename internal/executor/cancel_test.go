package executor_test

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/output"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWrapperMissingRealAWSCancelsWithStatus3(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapperDir := filepath.Join(work, "only-wrapper")

	err := os.MkdirAll(wrapperDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(wrapperDir, "aws"), "./cmd/aws")

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	secretArg := "SECRET_ARG_TOKEN_must_not_leak"
	cmd := exec.CommandContext(ctx, wrapper, "sts", secretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envWithout("PATH=" + wrapperDir)

	err = cmd.Run()

	if ctx.Err() != nil {
		t.Fatal("wrapper did not exit quickly when real AWS CLI is missing")
	}

	assertExitCode(t, err, output.CancelStatus)
	assertCancellationStderr(t, stderr.String(), secretArg)

	if strings.Contains(stderr.String(), executor.RealAWSEnv) {
		t.Fatalf("stderr must not mention %s", executor.RealAWSEnv)
	}
}

func TestWrapperExecFailureCancelsWithStatus3(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	marker := "path-marker-unique-abc123"
	missingReal := filepath.Join(work, marker, "no-such-aws-cli")

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")

	secretArg := "SECRET_ARG_TOKEN_must_not_leak"
	cmd := exec.CommandContext(t.Context(), wrapper, "sts", secretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envWithout(executor.RealAWSEnv + "=" + missingReal)

	err := cmd.Run()
	assertExitCode(t, err, output.CancelStatus)
	assertCancellationStderr(t, stderr.String(), secretArg)

	if strings.Contains(stderr.String(), marker) {
		t.Fatalf("stderr must not contain real CLI path marker %q", marker)
	}

	if strings.Contains(stderr.String(), missingReal) {
		t.Fatalf("stderr must not contain real CLI path")
	}
}

func TestWrapperSuccessfulHandoffWritesNoWrapperStderr(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envWithout(
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
	)

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty wrapper stderr on handoff, got %q", stderr.String())
	}
}

func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != want {
		t.Fatalf("expected exit %d, got %v", want, err)
	}
}

func assertCancellationStderr(t *testing.T, stderr, secretArg string) {
	t.Helper()

	if !strings.Contains(stderr, "aws-cli-guard: error: command was not executed") {
		t.Fatalf("missing cancellation line in stderr: %q", stderr)
	}

	if !strings.Contains(stderr, "aws-cli-guard: error: AWS CLI Guard stopped it:") {
		t.Fatalf("missing stopped-it line in stderr: %q", stderr)
	}

	if strings.Contains(stderr, secretArg) {
		t.Fatalf("stderr must not contain argv token")
	}

	if strings.Contains(stderr, "aws-cli-guard: warning:") {
		t.Fatalf("cancellation must not emit warning prefix: %q", stderr)
	}
}
