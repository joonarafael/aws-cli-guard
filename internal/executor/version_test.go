package executor_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapperOlderCLIHandoffWithoutPolicy(t *testing.T) {
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
			"FAKE_AWS_EXIT=17",
		),
		"sts",
		cancellationSecretArg,
	)
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 17 {
			t.Fatalf("run wrapper: %v", err)
		}
	} else {
		t.Fatal("expected exit 17")
	}

	wantWarn := "AWS CLI 2.8.9 is older than 2.9.0; running it with no policy"
	if !strings.Contains(stderr, wantWarn) {
		t.Fatalf("stderr = %q", stderr)
	}

	if strings.Contains(stderr, cancellationSecretArg) {
		t.Fatal("stderr leaked secret arg")
	}

	if strings.Contains(stderr, fake) {
		t.Fatal("stderr leaked fake path")
	}

	if strings.Contains(stderr, "aws-cli-guard: error:") {
		t.Fatalf("must not cancel: %q", stderr)
	}

	args := readRecordedArgs(t, filepath.Join(work, "record.json"))
	if args[1] != "sts" {
		t.Fatalf("args = %v", args)
	}
}

func TestWrapperOlderCLIWarnsOncePerSession(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	env := append(
		versionHandoffEnv(t, work, "aws-cli/2.8.9 Python/3.11.0"),
		"FAKE_AWS_EXIT=0",
	)

	_, err := runWrapperVersionCase(t, work, wrapper, fake, env, "s3", "ls")
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	if readVersionProbeCount(t, work) != 1 {
		t.Fatalf("probe count after first = %d", readVersionProbeCount(t, work))
	}

	stderr, err := runWrapperVersionCase(t, work, wrapper, fake, env, "s3", "ls")
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if readVersionProbeCount(t, work) != 1 {
		t.Fatal("second run should not probe again")
	}

	if strings.Contains(stderr, "older than 2.9.0") {
		t.Fatalf("second run should not warn again: %q", stderr)
	}
}

func TestWrapperUnreadableCLIHandoff(t *testing.T) {
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
		append(versionHandoffEnv(t, work, "unreadable"), "FAKE_AWS_EXIT=3"),
		"s3",
		"ls",
	)
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
			t.Fatalf("run: %v", err)
		}
	}

	if !strings.Contains(stderr, "version could not be read") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestWrapperNewerCLIStillGuards(t *testing.T) {
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
			emptyPolicyConfigEnvEntries(t, work),
			"FAKE_AWS_VERSION=aws-cli/2.9.1 Python/3.11.0",
		),
		"sts",
		cancellationSecretArg,
	)

	assertExitCode(t, err)
	assertCancellationWithOptionalWarning(t, stderr)

	if !strings.Contains(stderr, "newer than the verified release") {
		t.Fatalf("stderr = %q", stderr)
	}

	assertNoRecord(t, filepath.Join(work, "record.json"))
}

func TestWrapperNewerCLIWarnsOnceAndStillCancels(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	env := append(
		emptyPolicyConfigEnvEntries(t, work),
		"FAKE_AWS_VERSION=aws-cli/2.9.1 Python/3.11.0",
	)

	_, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		env,
		"sts",
		cancellationSecretArg,
	)
	assertExitCode(t, err)

	stderr, err := runWrapperVersionCase(
		t,
		work,
		wrapper,
		fake,
		env,
		"sts",
		cancellationSecretArg,
	)
	assertExitCode(t, err)

	if strings.Contains(stderr, "newer than the verified release") {
		t.Fatalf("second run should not warn again: %q", stderr)
	}

	if readVersionProbeCount(t, work) != 1 {
		t.Fatal("expected single probe")
	}
}
