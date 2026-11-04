//go:build unix

package agent

import (
	"errors"
	"os"
	"syscall"
)

func classifyStdin(stdin *os.File) Stdin {
	if stdin == nil {
		return StdinUndetermined
	}

	err := ioctlIsTerminal(int(stdin.Fd()))
	if err == nil {
		return StdinInteractive
	}

	if isNotTTYErr(err) {
		return StdinNonInteractive
	}

	return StdinUndetermined
}

func isNotTTYErr(err error) bool {
	var errno syscall.Errno

	if !errors.As(err, &errno) {
		return false
	}

	return errno == syscall.ENOTTY || errno == syscall.EINVAL || errno == syscall.ENXIO
}
