package executor_test

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/policy"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapperMissingPolicyCancelsWithStatus3(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", cancellationSecretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envHandoff(
		emptyPolicyConfigEnv(t, work),
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), policy.ReasonMissing) {
		t.Fatalf("stderr = %q", stderr.String())
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run when policy is missing")
	}
}

func TestWrapperPolicyPathUnsetCancelsWithStatus3(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", cancellationSecretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envHandoff(
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
	)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), policy.ReasonPathUnset) {
		t.Fatalf("stderr = %q", stderr.String())
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run when policy path is unset")
	}
}
