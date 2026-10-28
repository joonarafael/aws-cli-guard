package terminal

import "errors"

// ErrNoTerminal is returned when the controlling terminal cannot be opened.
var ErrNoTerminal = errors.New("aws-cli-guard: no controlling terminal")

// ErrMismatch is returned when the typed confirmation does not match the expected phrase.
var ErrMismatch = errors.New("aws-cli-guard: confirmation did not match")

// ErrInvalidPhrase is returned when phrase is empty or contains control characters.
var ErrInvalidPhrase = errors.New("aws-cli-guard: invalid confirmation phrase")

// ErrPlatformUnsupported is returned when terminal confirmation is unavailable on this GOOS.
var ErrPlatformUnsupported = errors.New(
	"aws-cli-guard: terminal confirmation is not supported on this platform yet",
)
