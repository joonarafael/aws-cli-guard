package executor_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aws-cli-guard/internal/executor"
)

func TestWrapperForwardsArgsAndExitStatus(t *testing.T) {
	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.Command(wrapper, "sts", "get-caller-identity", "--output", "json")
	cmd.Env = envWithout(executor.RealAWSEnv,
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"FAKE_AWS_EXIT=42",
	)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 42 {
			t.Fatalf("run wrapper: %v", err)
		}
	} else {
		t.Fatal("expected non-zero exit status 42")
	}

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	if len(args) < 5 {
		t.Fatalf("args too short: %v", args)
	}
	if args[1] != "sts" || args[2] != "get-caller-identity" || args[3] != "--output" || args[4] != "json" {
		t.Fatalf("unexpected forwarded args: %v", args)
	}
	if filepath.Base(args[0]) != "fake-aws" {
		t.Fatalf("expected argv0 basename fake-aws, got %q", args[0])
	}
}

func TestWrapperSkipsSelfOnPATH(t *testing.T) {
	root := moduleRoot(t)
	work := t.TempDir()

	wrapperDir := filepath.Join(work, "wrapper-bin")
	fakeDir := filepath.Join(work, "real-bin")
	if err := os.MkdirAll(wrapperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(fakeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wrapper := build(t, root, filepath.Join(wrapperDir, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(fakeDir, "aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.Command(wrapper, "s3", "ls")
	cmd.Env = envWithout(executor.RealAWSEnv,
		"PATH="+wrapperDir+string(os.PathListSeparator)+fakeDir,
		"FAKE_AWS_RECORD="+record,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
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

func TestWrapperMissingRealAWSExitsQuickly(t *testing.T) {
	root := moduleRoot(t)
	work := t.TempDir()

	wrapperDir := filepath.Join(work, "only-wrapper")
	if err := os.MkdirAll(wrapperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wrapper := build(t, root, filepath.Join(wrapperDir, "aws"), "./cmd/aws")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, wrapper, "sts", "get-caller-identity")
	cmd.Env = envWithout(executor.RealAWSEnv, "PATH="+wrapperDir)

	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("wrapper did not exit quickly when real AWS CLI is missing")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit 1, got %v", err)
	}
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
	cmd := exec.Command("go", "build", "-o", out, pkg)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, out)
	}
	return out
}

func envWithout(key string, extra ...string) []string {
	var env []string
	prefix := key + "="
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		env = append(env, e)
	}
	return append(env, extra...)
}
