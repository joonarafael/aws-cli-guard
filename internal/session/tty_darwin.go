//go:build darwin

package session

import (
	"syscall"
	"unsafe"
)

const (
	fGetPath      = 50
	darwinPathMax = 1024
)

func controllingTTYPath(fd int) (string, error) {
	buf := make([]byte, darwinPathMax)

	_, _, errno := syscall.Syscall(
		syscall.SYS_FCNTL,
		uintptr(fd),
		fGetPath,
		uintptr(unsafe.Pointer(&buf[0])), //nolint:gosec // fcntl buffer
	)
	if errno != 0 {
		return "", ErrNoTTY
	}

	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}

	return string(buf[:n]), nil
}

func statDev(stat *syscall.Stat_t) uint64 {
	//nolint:gosec // darwin st_dev is int32; preserve device bits.
	return uint64(uint32(stat.Dev))
}

func statIno(stat *syscall.Stat_t) uint64 {
	return stat.Ino
}

func statRdev(stat *syscall.Stat_t) uint64 {
	//nolint:gosec // darwin st_rdev is int32; preserve device bits.
	return uint64(uint32(stat.Rdev))
}
