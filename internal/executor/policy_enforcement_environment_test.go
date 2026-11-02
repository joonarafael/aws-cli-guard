package executor_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	envCredAccessKeyVar = "AWS_ACCESS_KEY_ID"
	envCredSecretVar    = "AWS_SECRET_ACCESS_KEY"
	envCredAccessValue  = "AKIAENVTESTKEY"
	envCredSecretValue  = "env-test-secret-value"
)

func TestWrapperEnvironmentDenyDoesNotExec(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, environmentPolicy("deny"))
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", "get-caller-identity")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	cmd.Env = append(
		enforceEnv(t, work, fake, record),
		envCredAccessKeyVar+"="+envCredAccessValue,
		envCredSecretVar+"="+envCredSecretValue,
	)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if strings.Contains(stderr.String(), envCredAccessValue) {
		t.Fatal("stderr must not echo the access key id")
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run when environment policy denies")
	}
}

func TestWrapperEnvironmentAllowExecsOnce(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, environmentPolicy("allow"))
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", "get-caller-identity")
	cmd.Env = append(
		enforceEnv(t, work, fake, record),
		envCredAccessKeyVar+"="+envCredAccessValue,
		envCredSecretVar+"="+envCredSecretValue,
	)

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	readRecordedArgs(t, record)
}

func environmentPolicy(action string) string {
	return `version = 2

[defaults]
action = "deny"

[environment]
action = "` + action + `"

[credential_export]
action = "allow"
`
}
