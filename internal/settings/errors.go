package settings

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrMissing is returned when the settings file does not exist.
var ErrMissing = errors.New("aws-cli-guard: settings: file missing")

// ErrTooLarge is returned when the settings file exceeds the size limit.
var ErrTooLarge = errors.New("aws-cli-guard: settings: file too large")

// Kind classifies a settings load or write failure.
type Kind int

// Failure kinds returned in settings.Error.
const (
	KindMissing Kind = iota + 1
	KindRead
	KindPath
	KindDecode
	KindInvalid
	KindWrite
)

// Error is a secret-safe settings failure.
type Error struct {
	Kind       Kind
	Path       string
	Key        string
	Line       int
	Column     int
	Version    int
	HasVersion bool
	Supported  []int
	wrapped    error
}

func newError(kind Kind, path, key string, line, column int, wrapped error) *Error {
	return &Error{
		Kind:       kind,
		Path:       stripControls(path),
		Key:        stripControls(key),
		Line:       line,
		Column:     column,
		Version:    0,
		HasVersion: false,
		Supported:  nil,
		wrapped:    wrapped,
	}
}

func (e *Error) Error() string {
	var builder strings.Builder

	switch e.Kind {
	case KindMissing:
		builder.WriteString("aws-cli-guard: settings: file missing")
	case KindRead:
		builder.WriteString("aws-cli-guard: settings: read failed")
	case KindPath:
		builder.WriteString("aws-cli-guard: settings: invalid path")
	case KindDecode:
		builder.WriteString("aws-cli-guard: settings: invalid document")
	case KindInvalid:
		builder.WriteString("aws-cli-guard: settings: invalid value")
	case KindWrite:
		builder.WriteString("aws-cli-guard: settings: write failed")
	default:
		builder.WriteString("aws-cli-guard: settings: error")
	}

	if e.Path != "" {
		fmt.Fprintf(&builder, " in %q", e.Path)
	}

	if e.Key != "" {
		fmt.Fprintf(&builder, " at key %q", e.Key)
	}

	if e.Line > 0 {
		fmt.Fprintf(&builder, " at line %d column %d", e.Line, e.Column)
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
