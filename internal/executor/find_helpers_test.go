package executor_test

import (
	"aws-cli-guard/internal/executor"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func copyWrapperBinary(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(dst, data, 0o755)
	if err != nil {
		t.Fatal(err)
	}
}

func assertResolvedSameFile(t *testing.T, leftPath, rightPath string) {
	t.Helper()

	resolvedLeft, err := filepath.EvalSymlinks(leftPath)
	if err != nil {
		resolvedLeft = leftPath
	}

	resolvedRight, err := filepath.EvalSymlinks(rightPath)
	if err != nil {
		resolvedRight = rightPath
	}

	stLeft, err := os.Stat(resolvedLeft)
	if err != nil {
		t.Fatalf("stat left: %v", err)
	}

	stRight, err := os.Stat(resolvedRight)
	if err != nil {
		t.Fatalf("stat right: %v", err)
	}

	if !os.SameFile(stLeft, stRight) {
		t.Fatalf("paths differ: %q vs %q", leftPath, rightPath)
	}
}

func assertNoRecord(t *testing.T, record string) {
	t.Helper()

	_, err := os.Stat(record)
	if err == nil {
		t.Fatalf("record file %q should not exist", record)
	}
}

func assertStderrOmitsRealAWSConfig(t *testing.T, stderr, configuredPath string) {
	t.Helper()

	if strings.Contains(stderr, executor.RealAWSEnv) {
		t.Fatalf("stderr must not mention %s", executor.RealAWSEnv)
	}

	if strings.Contains(stderr, configuredPath) {
		t.Fatalf("stderr must not contain configured path %q", configuredPath)
	}
}
