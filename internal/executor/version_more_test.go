package executor_test

import (
	"aws-cli-guard/internal/executor"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapperSupportedDefaultNoVersionWarning(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, denyProfilePolicy())
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")

	stderr, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		guardDirEnvEntries(t, work),
		"sts",
		"get-caller-identity",
		"--profile",
		enforceProfile,
	)

	assertExitCode(t, err)

	if strings.Contains(stderr, "aws-cli-guard: warning:") {
		t.Fatalf("unexpected warning: %q", stderr)
	}

	assertNoRecord(t, filepath.Join(work, "record.json"))
}

func TestWrapperCorruptCLIVersionCacheStillGuards(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeCorruptCLIVersionCache(t, filepath.Join(work, "empty-guard-cache"))
	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")

	stderr, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		emptyPolicyConfigEnvEntries(t, work),
		"sts",
		cancellationSecretArg,
	)

	assertExitCode(t, err)

	if !strings.Contains(stderr, "cache file") || !strings.Contains(stderr, "unusable") {
		t.Fatalf("stderr = %q", stderr)
	}

	before, err := os.ReadFile(filepath.Join(work, "empty-guard-cache", "cli-version.json"))
	if err != nil {
		t.Fatal(err)
	}

	if string(before) != "{bad" {
		t.Fatal("corrupt cache was modified")
	}

	assertNoRecord(t, filepath.Join(work, "record.json"))
}

func TestWrapperCLIVersionProbeUsesStderr(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")

	_, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		append(
			versionHandoffEnv(t, work, "omit"),
			"FAKE_AWS_VERSION_STDERR=aws-cli/2.9.0 Python/3.11.0",
		),
		"s3",
		"ls",
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if readVersionProbeCount(t, work) != 1 {
		t.Fatal("expected version probe")
	}
}

func TestWrapperStdinForwardedOnGuardedExec(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	stdinCapture := filepath.Join(work, "stdin.txt")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_STDIN="+stdinCapture,
	)...)
	cmd.Stdin = strings.NewReader("payload-bytes")

	err := cmd.Run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	got, err := os.ReadFile(stdinCapture)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "payload-bytes" {
		t.Fatalf("stdin = %q", got)
	}
}
