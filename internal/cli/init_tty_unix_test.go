//go:build unix

package cli_test

import (
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const initHeadlessChildEnv = "AWS_CLI_GUARD_INIT_HEADLESS_CHILD"

func TestHeadlessInitIgnoresControllingTerminal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(moduleRootForCLIDisable(t), "scripts", "pty-must-exit.py")

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", script, os.Args[0], "-test.run=^$")

	cmd.Env = append(
		os.Environ(),
		initHeadlessChildEnv+"=1",
		"AWS_CLI_GUARD_CONFIG_DIR="+dir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("headless init under a pty: %v\n%s", err, output)
	}

	assertInitDefaults(t, dir)
}

func assertInitDefaults(t *testing.T, dir string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatal("policy body mismatch")
	}

	data, err = os.ReadFile(filepath.Join(dir, "settings.toml"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != settings.DefaultFileBody {
		t.Fatalf("settings body = %q", data)
	}
}

func runInitHeadlessChild() int {
	dir := os.Getenv("AWS_CLI_GUARD_CONFIG_DIR")

	var stderr bytes.Buffer

	status := runHeadlessInit(
		[]string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir},
		&bytes.Buffer{},
		&stderr,
	)
	if status != 0 {
		_, _ = os.Stderr.Write(stderr.Bytes())
	}

	return status
}
