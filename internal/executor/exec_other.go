//go:build !unix

package executor

import (
	"fmt"
	"os"
)

// ExecRealAWS replaces the current process with the real AWS CLI.
func ExecRealAWS(realPath string, args []string, env []string) error {
	return fmt.Errorf("aws-cli-guard: process replacement is not supported on this platform yet")
}

// Environ returns the process environment for exec.
func Environ() []string {
	return os.Environ()
}
