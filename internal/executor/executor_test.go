package executor_test

import (
	"aws-cli-guard/internal/executor"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestWrapperForwardsArgsAndExitStatus(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"sts",
		"get-caller-identity",
		"--output",
		"json",
	)
	cmd.Env = envHandoff(
		policyConfigEnv(t, work),
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"FAKE_AWS_EXIT=42",
		"AWS_PROFILE="+handoffProfileName,
	)

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 42 {
			t.Fatalf("run wrapper: %v", err)
		}
	} else {
		t.Fatal("expected non-zero exit status 42")
	}

	args := readRecordedArgs(t, record)

	if len(args) < 5 {
		t.Fatalf("args too short: %v", args)
	}

	if args[1] != "sts" || args[2] != "get-caller-identity" || args[3] != "--output" ||
		args[4] != "json" {
		t.Fatalf("unexpected forwarded args: %v", args)
	}

	if filepath.Base(args[0]) != "fake-aws" {
		t.Fatalf("expected argv0 basename fake-aws, got %q", args[0])
	}
}

func TestWrapperSkipsSelfOnPATH(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapperDir := filepath.Join(work, "wrapper-bin")
	fakeDir := filepath.Join(work, "real-bin")

	err := os.MkdirAll(wrapperDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(fakeDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(wrapperDir, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(t.Context(), wrapper, "s3", "ls")

	cmd.Env = envHandoff(
		policyConfigEnv(t, work),
		"PATH="+wrapperDir+string(os.PathListSeparator)+fakeDir,
		"FAKE_AWS_RECORD="+record,
		"AWS_PROFILE="+handoffProfileName,
	)

	err = cmd.Run()
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	args := readRecordedArgs(t, record)

	resolvedArg0, err := filepath.EvalSymlinks(args[0])
	if err != nil {
		resolvedArg0 = args[0]
	}

	resolvedFake, err := filepath.EvalSymlinks(fake)
	if err != nil {
		resolvedFake = fake
	}

	if resolvedArg0 != resolvedFake {
		t.Fatalf("expected fake aws on PATH after wrapper, argv0=%q want %q", args[0], fake)
	}

	if args[1] != "s3" || args[2] != "ls" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func readRecordedArgs(t *testing.T, record string) []string {
	t.Helper()

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	var args []string

	err = json.Unmarshal(data, &args)
	if err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}

	return args
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func build(t *testing.T, dir, out, pkg string) string {
	t.Helper()

	buildArgs := []string{"build", "-o", out, pkg}
	if raceDetectorActive() {
		buildArgs = []string{"build", "-race", "-o", out, pkg}
	}

	cmd := exec.CommandContext(t.Context(), "go", buildArgs...)
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, output)
	}

	return out
}

func raceDetectorActive() bool {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}

	for _, setting := range info.Settings {
		if setting.Key == "-race" && setting.Value == "true" {
			return true
		}
	}

	return false
}

const handoffProfileName = "staging-handoff-test"

func envWithout(extra ...string) []string {
	var env []string

	prefix := executor.RealAWSEnv + "="
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, prefix) {
			continue
		}

		env = append(env, e)
	}

	return append(env, extra...)
}

func envHandoff(extra ...string) []string {
	base := envWithout()

	var cleaned []string

	stripPrefixes := []string{
		"AWS_ACCESS_KEY_ID=",
		"AWS_SECRET_ACCESS_KEY=",
		"AWS_SESSION_TOKEN=",
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI=",
		"AWS_CONTAINER_CREDENTIALS_FULL_URI=",
		"AWS_WEB_IDENTITY_TOKEN_FILE=",
		"AWS_PROFILE=",
		"AWS_CLI_GUARD_CONFIG_DIR=",
		"XDG_CONFIG_HOME=",
		"HOME=",
	}

	for _, entry := range base {
		skip := false

		for _, prefix := range stripPrefixes {
			if strings.HasPrefix(entry, prefix) {
				skip = true

				break
			}
		}

		if !skip {
			cleaned = append(cleaned, entry)
		}
	}

	return append(cleaned, extra...)
}
