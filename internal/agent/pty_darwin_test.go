//go:build darwin

//nolint:testpackage // asserts Observe results through unexported stdin field.
package agent

import (
	"os"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestObservePTYInteractive(t *testing.T) {
	t.Parallel()

	slave := openPTYSlaveDarwin(t)

	ctx := Observe(slave)
	if ctx.stdin != StdinInteractive {
		t.Fatalf("stdin = %v, want interactive", ctx.stdin)
	}
}

func openPTYSlaveDarwin(t *testing.T) *os.File {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = master.Close() })

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCPTYGRANT, 0)
	if errno != 0 {
		t.Fatal(errno)
	}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCPTYUNLK, 0)
	if errno != 0 {
		t.Fatal(errno)
	}

	var name [256]byte

	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCPTYGNAME,
		uintptr(unsafe.Pointer(&name[0])),
	)
	if errno != 0 {
		t.Fatal(errno)
	}

	slavePath := "/dev/" + strings.TrimRight(string(name[:]), "\x00")

	slave, err := os.OpenFile(slavePath, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = slave.Close() })

	return slave
}
