//go:build unix

package executor

import (
	"fmt"
	"os"
	"syscall"
)

// ExecRealAWS replaces the current process with the real AWS CLI.
func ExecRealAWS(realPath string, args []string, env []string) error {
	execArgv := append([]string{realPath}, args[1:]...)
	if err := syscall.Exec(realPath, execArgv, env); err != nil {
		return fmt.Errorf("aws-cli-guard: failed to exec real AWS CLI: %w", err)
	}
	return nil
}

// Environ returns the process environment for exec.
func Environ() []string {
	return os.Environ()
}
