//go:build unix

package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// StartDetachedCheck starts aws-cli-guard-update-check in a new session.
func StartDetachedCheck(executable string, env []string) error {
	cmd := exec.CommandContext(context.Background(), executable, CheckArgv)
	cmd.Env = env
	cmd.Stdin = devNull()
	cmd.Stdout = devNull()
	cmd.Stderr = devNull()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("aws-cli-guard: update: start helper: %w", err)
	}

	return nil
}

func devNull() *os.File {
	file, err := os.Open(os.DevNull)
	if err != nil {
		return nil
	}

	return file
}
