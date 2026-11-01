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

const (
	enforceProfile     = "production-admin"
	enforceGrantExpiry = "2099-01-01T00:00:00Z"
)

func TestWrapperDenyDoesNotExec(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, denyProfilePolicy())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"sts",
		"get-caller-identity",
		"--profile",
		enforceProfile,
		cancellationSecretArg,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run when policy denies")
	}
}

func TestWrapperPromptWithoutGrantDoesNotExec(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, promptProfilePolicy())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"ec2",
		"describe-instances",
		"--profile",
		enforceProfile,
		cancellationSecretArg,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), "aws-guard allow "+enforceProfile) {
		t.Fatalf("stderr = %q", stderr.String())
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run without grant")
	}
}

func TestWrapperPromptWithGrantExecsOnce(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, promptProfilePolicy())
	writeEnforceGrants(t, work, currentGrantJSON())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"s3",
		"ls",
		"--profile",
		enforceProfile,
	)

	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	readRecordedArgs(t, record)
}

func TestWrapperExportBlockedDoesNotExec(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, exportPromptAllowProfilePolicy())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"configure",
		"export-credentials",
		"--profile",
		enforceProfile,
		cancellationSecretArg,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), policy.ReasonExportRefused) {
		t.Fatalf("stderr = %q", stderr.String())
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run for blocked export")
	}
}

func TestWrapperCorruptPolicyUnchanged(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	configDir := filepath.Join(work, "guard-config")
	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("not valid toml\n")
	policyPath := filepath.Join(configDir, "config.toml")

	err = os.WriteFile(policyPath, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", cancellationSecretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envHandoff(append(
		[]string{
			"AWS_CLI_GUARD_CONFIG_DIR=" + configDir,
			"AWS_CLI_GUARD_STATE_DIR=" + stateDir,
		},
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err = cmd.Run()
	assertExitCode(t, err)

	after, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(after) != string(body) {
		t.Fatal("corrupt policy bytes changed")
	}
}
