package settings

import (
	"fmt"
	"strings"
	"unicode"
)

// Kind classifies a settings create failure.
type Kind int

// Failure kinds returned in settings.Error.
const (
	KindPath Kind = iota + 1
	KindWrite
)

// Error is a secret-safe settings failure.
type Error struct {
	Kind    Kind
	Path    string
	wrapped error
}

func newError(kind Kind, path string, wrapped error) *Error {
	return &Error{
		Kind:    kind,
		Path:    stripControls(path),
		wrapped: wrapped,
	}
}

func (e *Error) Error() string {
	var builder strings.Builder

	switch e.Kind {
	case KindPath:
		builder.WriteString("aws-cli-guard: settings: invalid path")
	case KindWrite:
		builder.WriteString("aws-cli-guard: settings: write failed")
	default:
		builder.WriteString("aws-cli-guard: settings: error")
	}

	if e.Path != "" {
		fmt.Fprintf(&builder, " in %q", e.Path)
	}

	return builder.String()
}

func (e *Error) Unwrap() error {
	return e.wrapped
}

func stripControls(text string) string {
	return strings.Map(func(runeVal rune) rune {
		if runeVal == unicode.ReplacementChar {
			return runeVal
		}

		if unicode.IsControl(runeVal) {
			return -1
		}

		return runeVal
	}, text)
}
