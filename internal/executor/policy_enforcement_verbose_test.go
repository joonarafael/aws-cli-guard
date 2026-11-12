package executor_test

import (
	"aws-cli-guard/internal/executor"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapperVerboseAllow(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, allowProfilePolicy())
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
		cancellationSecretArg,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	cmd.Env = append(enforceEnv(t, work, fake, record), "AWS_CLI_GUARD_LOG=verbose")

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v, stderr=%q", err, stderr.String())
	}

	out := stderr.String()
	if !strings.Contains(out, "outcome: allow") || !strings.Contains(out, "aws: "+fake) {
		t.Fatalf("stderr = %q", out)
	}

	if strings.Contains(out, cancellationSecretArg) {
		t.Fatal("stderr leaked secret arg")
	}
}

func TestWrapperTightAllowNoOutcomeLine(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, allowProfilePolicy())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls", "--profile", enforceProfile)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	if strings.Contains(stderr.String(), "outcome:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestWrapperVerboseSSOLoginHandoff(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, execSSOCommand, execLoginOp)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	home := filepath.Join(work, "home")

	err := os.MkdirAll(home, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	cmd.Env = envHandoff(append(
		guardDirEnvEntries(t, work),
		"HOME="+home,
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"AWS_CLI_GUARD_LOG=verbose",
	)...)

	runErr := cmd.Run()
	if runErr != nil {
		t.Fatalf("run wrapper: %v", runErr)
	}

	out := stderr.String()
	if !strings.Contains(out, "outcome: login handoff") {
		t.Fatalf("stderr = %q", out)
	}
}

func TestWrapperVerboseOlderCLIVersionHandoff(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")

	stderr, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		append(
			guardDirEnvEntries(t, work),
			"FAKE_AWS_VERSION=aws-cli/2.8.9 Python/3.11.0",
			"FAKE_AWS_EXIT=0",
			"AWS_CLI_GUARD_LOG=verbose",
		),
		"s3",
		"ls",
	)
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	if !strings.Contains(stderr, "outcome: version handoff") {
		t.Fatalf("stderr = %q", stderr)
	}
}
