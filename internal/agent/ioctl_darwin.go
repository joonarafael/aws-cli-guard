//go:build darwin

package agent

import (
	"syscall"
	"unsafe"
)

func ioctlIsTerminal(fd int) error {
	var term syscall.Termios

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		syscall.TIOCGETA,
		uintptr(unsafe.Pointer(&term)), //nolint:gosec // ioctl requires termios address
	)
	if errno != 0 {
		return errno
	}

	return nil
}
