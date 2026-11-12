//go:build unix

package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardDisableExecsSiblingStub(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	stub := filepath.Join(dir, "aws-cli-guard-disable")

	err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	root := moduleRootForCLIDisable(t)
	guard := filepath.Join(dir, "aws-guard")
	buildGuard(t, root, guard)

	cmd := exec.CommandContext(t.Context(), guard, "disable")

	err = cmd.Run()
	if err != nil {
		t.Fatalf("guard disable: %v", err)
	}
}

func moduleRootForCLIDisable(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func buildGuard(t *testing.T, root, out string) {
	t.Helper()

	cmd := exec.CommandContext(
		t.Context(),
		"go", "build", "-tags=awsguard_testpaths", "-o", out, "./cmd/aws-guard",
	)
	cmd.Dir = root

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build guard: %v\n%s", err, output)
	}
}

func TestGuardDisableRequiresSibling(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	guard := filepath.Join(dir, "aws-guard")

	root := moduleRootForCLIDisable(t)
	buildGuard(t, root, guard)

	cmd := exec.CommandContext(t.Context(), guard, "disable")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected failure without sibling script")
	}

	if !strings.Contains(stderr.String(), "the disable script is not available") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
