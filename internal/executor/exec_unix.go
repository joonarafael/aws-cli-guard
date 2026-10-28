//go:build unix

package executor

import (
	"os"
	"syscall"
)

// ExecRealAWS replaces the current process with the real AWS CLI.
func ExecRealAWS(realPath string, args []string, env []string) error {
	execArgv := append([]string{realPath}, args[1:]...)

	//nolint:gosec // G204: deliberate exec of the resolved real AWS CLI.
	err := syscall.Exec(realPath, execArgv, env)
	if err != nil {
		return ErrStartFailed
	}

	return nil
}

// Environ returns the process environment for exec.
func Environ() []string {
	return os.Environ()
}
