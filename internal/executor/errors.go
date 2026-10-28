package executor

import "errors"

// ErrRealAWSNotFound is returned when no suitable real AWS CLI executable is found on PATH.
var ErrRealAWSNotFound = errors.New("aws-cli-guard: real AWS CLI not found in PATH")

// ErrResolveWrapper is returned when the wrapper executable path cannot be resolved.
var ErrResolveWrapper = errors.New("aws-cli-guard: cannot resolve wrapper executable")

// ErrStartFailed is returned when the real AWS CLI could not be started.
var ErrStartFailed = errors.New("aws-cli-guard: failed to start real AWS CLI")

// ErrPlatformUnsupported is returned when process replacement is unavailable on this GOOS.
var ErrPlatformUnsupported = errors.New(
	"aws-cli-guard: process replacement is not supported on this platform yet",
)

const (
	reasonRealAWSNotFound     = "the real AWS CLI was not found"
	reasonResolveWrapper      = "the wrapper executable could not be resolved"
	reasonStartFailed         = "the real AWS CLI could not be started"
	reasonPlatformUnsupported = "process replacement is not supported on this platform"
	reasonGeneric             = "the command could not be started"
)

// CancellationReason maps a wrapper cancellation error to a fixed, secret-safe reason line.
func CancellationReason(err error) string {
	switch {
	case errors.Is(err, ErrRealAWSNotFound):
		return reasonRealAWSNotFound
	case errors.Is(err, ErrResolveWrapper):
		return reasonResolveWrapper
	case errors.Is(err, ErrStartFailed):
		return reasonStartFailed
	case errors.Is(err, ErrPlatformUnsupported):
		return reasonPlatformUnsupported
	default:
		return reasonGeneric
	}
}
