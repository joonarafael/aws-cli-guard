//go:build unix

package session

import (
	"aws-cli-guard/internal/sessions"
	"syscall"
)

const devTTY = "/dev/tty"

// CaptureControllingTTY opens the controlling terminal and records its device identity.
func CaptureControllingTTY() (sessions.TTY, error) {
	ttyFD, err := syscall.Open(devTTY, syscall.O_RDONLY|syscall.O_NOCTTY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return sessions.TTY{}, ErrNoTTY
	}

	defer func() {
		_ = syscall.Close(ttyFD)
	}()

	var stat syscall.Stat_t

	err = syscall.Fstat(ttyFD, &stat)
	if err != nil {
		return sessions.TTY{}, ErrNoTTY
	}

	path, pathErr := controllingTTYPath(ttyFD)
	if pathErr != nil {
		return sessions.TTY{}, pathErr
	}

	return sessions.TTY{
		Path: path,
		Dev:  statDev(&stat),
		Ino:  statIno(&stat),
		Rdev: statRdev(&stat),
	}, nil
}

// CurrentControllingTTY reports the controlling terminal when this process has one.
func CurrentControllingTTY() (sessions.TTY, bool) {
	tty, err := CaptureControllingTTY()
	if err != nil {
		return sessions.TTY{}, false //nolint:exhaustruct_v5 // zero value means absent.
	}

	return tty, true
}
