package executor

import "errors"

// ErrRealAWSNotFound is returned when no suitable real AWS CLI executable is found on PATH.
var ErrRealAWSNotFound = errors.New("aws-cli-guard: real AWS CLI not found in PATH")

// ErrPlatformUnsupported is returned when process replacement is unavailable on this GOOS.
var ErrPlatformUnsupported = errors.New(
	"aws-cli-guard: process replacement is not supported on this platform yet",
)
