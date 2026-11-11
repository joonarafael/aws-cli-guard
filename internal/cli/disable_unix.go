//go:build unix

package cli

import (
	"aws-cli-guard/internal/output"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const reasonDisableScriptMissing = "the disable script is not available in this package"

func runPersistentDisableScript(action string, stdout, stderr io.Writer) int {
	exe, err := os.Executable()
	if err != nil {
		output.Error(stderr, reasonDisableScriptMissing)

		return output.CancelStatus
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		output.Error(stderr, reasonDisableScriptMissing)

		return output.CancelStatus
	}

	script := filepath.Join(filepath.Dir(exe), "aws-cli-guard-disable")

	info, err := os.Stat(script)
	if err != nil || info.IsDir() {
		output.Error(stderr, reasonDisableScriptMissing)

		return output.CancelStatus
	}

	//nolint:gosec // G204: script path is the sibling beside the running aws-guard binary.
	cmd := exec.CommandContext(context.Background(), script, action)
	cmd.Dir = filepath.Dir(script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}

	output.Error(stderr, reasonDisableScriptMissing)

	return output.CancelStatus
}
