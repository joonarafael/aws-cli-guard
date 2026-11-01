package executor_test

import (
	"aws-cli-guard/internal/executor"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const configuredWrapperReason = "the configured AWS CLI is an AWS CLI Guard wrapper"

func TestWrapperSkipsEarlierCopyOnPATH(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	copyDir := filepath.Join(work, "copy-bin")
	wrapperDir := filepath.Join(work, "run-bin")
	fakeDir := filepath.Join(work, "real-bin")

	for _, dir := range []string{copyDir, wrapperDir, fakeDir} {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			t.Fatal(err)
		}
	}

	wrapper := build(t, root, filepath.Join(wrapperDir, "aws"), "./cmd/aws")
	copyWrapperBinary(t, wrapper, filepath.Join(copyDir, "aws"))
	fake := build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, wrapper, "s3", "ls")
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		"PATH="+copyDir+string(os.PathListSeparator)+wrapperDir+
			string(os.PathListSeparator)+fakeDir,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err := cmd.Run()

	if ctx.Err() != nil {
		t.Fatal("wrapper did not exit quickly; possible wrapper handoff loop")
	}

	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	args := readRecordedArgs(t, record)
	assertResolvedSameFile(t, args[0], fake)
}

func TestWrapperConfiguredRealAWSNamesWrapperCancels(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	fakeDir := filepath.Join(work, "real-bin")

	err := os.MkdirAll(fakeDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		executor.RealAWSEnv+"="+wrapper,
		"PATH="+fakeDir,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err = cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), configuredWrapperReason) {
		t.Fatalf("stderr missing configured-wrapper reason: %q", stderr.String())
	}

	assertNoRecord(t, record)
	assertStderrOmitsRealAWSConfig(t, stderr.String(), wrapper)
}

func TestWrapperConfiguredRealAWSNamesCopyCancels(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	copyDir := filepath.Join(work, "copy-bin")
	fakeDir := filepath.Join(work, "real-bin")

	for _, dir := range []string{copyDir, fakeDir} {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			t.Fatal(err)
		}
	}

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	copyPath := filepath.Join(copyDir, "aws")
	copyWrapperBinary(t, wrapper, copyPath)
	build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		executor.RealAWSEnv+"="+copyPath,
		"PATH="+fakeDir,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err := cmd.Run()
	assertExitCode(t, err)
	assertCancellationStderr(t, stderr.String())

	if !strings.Contains(stderr.String(), configuredWrapperReason) {
		t.Fatalf("stderr missing configured-wrapper reason: %q", stderr.String())
	}

	assertNoRecord(t, record)
	assertStderrOmitsRealAWSConfig(t, stderr.String(), copyPath)
}

func TestWrapperFindsFakeAWSWhenNotOnPATH(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	fakeDir := filepath.Join(work, "real-bin")

	err := os.MkdirAll(fakeDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		"PATH="+fakeDir,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err = cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	args := readRecordedArgs(t, record)
	assertResolvedSameFile(t, args[0], fake)
}

func TestWrapperDoesNotSkipNonGoCandidateOnPATH(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	scriptDir := filepath.Join(work, "script-bin")
	fakeDir := filepath.Join(work, "real-bin")

	for _, dir := range []string{scriptDir, fakeDir} {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			t.Fatal(err)
		}
	}

	scriptAWS := filepath.Join(scriptDir, "aws")
	scriptBody := "#!/bin/sh\n: > \"$SCRIPT_AWS_MARKER\"\nexit 0\n"

	err := os.WriteFile(scriptAWS, []byte(scriptBody), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	marker := filepath.Join(work, "script-ran")
	record := filepath.Join(work, "record.json")

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")
	cmd.Env = envHandoff(append(
		policyConfigEnvEntries(t, work),
		"PATH="+scriptDir+string(os.PathListSeparator)+fakeDir,
		"SCRIPT_AWS_MARKER="+marker,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)...)

	err = cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	_, err = os.Stat(marker)
	if err != nil {
		t.Fatalf("expected script aws to run, marker missing: %v", err)
	}

	assertNoRecord(t, record)
}
