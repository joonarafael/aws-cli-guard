//go:build linux

//nolint:testpackage // asserts Observe results through unexported stdin field.
package agent

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"unsafe"
)

func TestObservePTYInteractive(t *testing.T) {
	t.Parallel()

	slave := openPTYSlaveLinux(t)

	ctx := Observe(slave)
	if ctx.stdin != StdinInteractive {
		t.Fatalf("stdin = %v, want interactive", ctx.stdin)
	}
}

func openPTYSlaveLinux(t *testing.T) *os.File {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = master.Close() })

	var unlock int32

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCSPTLCK,
		uintptr(unsafe.Pointer(&unlock)),
	)
	if errno != 0 {
		t.Fatal(errno)
	}

	var ptyNum int32

	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCGPTN,
		uintptr(unsafe.Pointer(&ptyNum)),
	)
	if errno != 0 {
		t.Fatal(errno)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", ptyNum)

	slave, err := os.OpenFile(slavePath, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = slave.Close() })

	return slave
}
