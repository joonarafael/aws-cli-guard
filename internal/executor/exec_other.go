//go:build !unix

package executor

import (
	"os"
)

// ExecRealAWS replaces the current process with the real AWS CLI.
func ExecRealAWS(_ string, _ []string, _ []string) error {
	return ErrPlatformUnsupported
}

// Environ returns the process environment for exec.
func Environ() []string {
	return os.Environ()
}
