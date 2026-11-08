//go:build !unix

package session

import (
	"aws-cli-guard/internal/sessions"
	"errors"
)

// ErrNoTTY means this process has no controlling terminal.
var ErrNoTTY = errors.New("aws-cli-guard: session: no controlling terminal")

// CaptureControllingTTY is unsupported on this platform.
func CaptureControllingTTY() (sessions.TTY, error) {
	return sessions.TTY{}, ErrNoTTY
}

// CurrentControllingTTY is unsupported on this platform.
func CurrentControllingTTY() (sessions.TTY, bool) {
	return sessions.TTY{}, false //nolint:exhaustruct_v5 // zero value means absent.
}
