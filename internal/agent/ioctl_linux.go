//go:build linux

package agent

import (
	"syscall"
	"unsafe"
)

const tcgets = 0x5401

func ioctlIsTerminal(fd int) error {
	var term syscall.Termios

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		tcgets,
		uintptr(unsafe.Pointer(&term)), //nolint:gosec // ioctl requires termios address
	)
	if errno != 0 {
		return errno
	}

	return nil
}
