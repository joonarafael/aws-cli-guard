//go:build unix

package session

import (
	"fmt"
	"os"
	"syscall"
)

// ID returns a stable key for the current terminal session.
func ID() string {
	sid, err := getsid(0)
	if err == nil && sid > 0 {
		return fmt.Sprintf("sid:%d", sid)
	}

	return fmt.Sprintf("ppid:%d", os.Getppid())
}

func getsid(pid int) (int, error) {
	ret, _, errno := syscall.Syscall(syscall.SYS_GETSID, uintptr(pid), 0, 0)
	if errno != 0 {
		return 0, errno
	}

	return int(ret), nil
}
