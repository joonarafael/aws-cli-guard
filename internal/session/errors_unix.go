//go:build unix

package session

import "errors"

// ErrNoTTY means this process has no controlling terminal.
var ErrNoTTY = errors.New("aws-cli-guard: session: no controlling terminal")
