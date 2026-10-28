package awsconfig

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrHomeUnset is returned when HOME is required for a default AWS file path.
var ErrHomeUnset = errors.New("aws-cli-guard: awsconfig: HOME is unset")

// ErrTildeUser is returned for ~user paths other than ~ or ~/.
var ErrTildeUser = errors.New("aws-cli-guard: awsconfig: unsupported ~user path")

// ErrNotRegular is returned when a path exists but is not a regular file.
var ErrNotRegular = errors.New("aws-cli-guard: awsconfig: path is not a regular file")

// ErrTooLarge is returned when a config file exceeds the size limit.
var ErrTooLarge = errors.New("aws-cli-guard: awsconfig: file too large")

var (
	errInvalidUTF8             = errors.New("aws-cli-guard: awsconfig: invalid UTF-8")
	errContinuationWithoutOpt  = errors.New("aws-cli-guard: awsconfig: continuation without option")
	errInvalidSectionHeader    = errors.New("aws-cli-guard: awsconfig: invalid section header")
	errDuplicateSection        = errors.New("aws-cli-guard: awsconfig: duplicate section")
	errInvalidOption           = errors.New("aws-cli-guard: awsconfig: invalid option")
	errOptionOutsideSection    = errors.New("aws-cli-guard: awsconfig: option outside section")
	errDuplicateKey            = errors.New("aws-cli-guard: awsconfig: duplicate key")
	errNestedLineWithoutEquals = errors.New("aws-cli-guard: awsconfig: nested line without =")
	errEmptyNestedKey          = errors.New("aws-cli-guard: awsconfig: empty nested key")
	errDuplicateNestedKey      = errors.New("aws-cli-guard: awsconfig: duplicate nested key")
)

// Kind classifies a read or parse failure.
type Kind int

// Failure kinds returned in awsconfig.Error.
const (
	KindParse Kind = iota + 1
	KindRead
	KindPath
)

// Error is a secret-safe awsconfig failure.
type Error struct {
	Kind    Kind
	Path    string
	Line    int
	Key     string
	wrapped error
}

func newError(kind Kind, path string, line int, key string, wrapped error) *Error {
	return &Error{
		Kind:    kind,
		Path:    path,
		Line:    line,
		Key:     stripControls(key),
		wrapped: wrapped,
	}
}

func (e *Error) Error() string {
	var builder strings.Builder

	switch e.Kind {
	case KindParse:
		builder.WriteString("aws-cli-guard: awsconfig: invalid syntax")
	case KindRead:
		builder.WriteString("aws-cli-guard: awsconfig: read failed")
	case KindPath:
		builder.WriteString("aws-cli-guard: awsconfig: invalid path")
	default:
		builder.WriteString("aws-cli-guard: awsconfig: error")
	}

	if e.Path != "" {
		fmt.Fprintf(&builder, " in %q", e.Path)
	}

	if e.Key != "" {
		fmt.Fprintf(&builder, " at key %q", e.Key)
	}

	if e.Line > 0 {
		fmt.Fprintf(&builder, " at line %d", e.Line)
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
