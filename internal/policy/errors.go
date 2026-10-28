package policy

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrMissing is returned when the policy file does not exist.
var ErrMissing = errors.New("aws-cli-guard: policy: file missing")

// ErrTooLarge is returned when the policy file exceeds the size limit.
var ErrTooLarge = errors.New("aws-cli-guard: policy: file too large")

// Kind classifies a policy load failure.
type Kind int

// Failure kinds returned in policy.Error.
const (
	KindMissing Kind = iota + 1
	KindRead
	KindPath
	KindDecode
	KindInvalid
)

// Error is a secret-safe policy load failure.
type Error struct {
	Kind    Kind
	Path    string
	Key     string
	Line    int
	Column  int
	wrapped error
}

func newError(kind Kind, path, key string, line, column int, wrapped error) *Error {
	return &Error{
		Kind:    kind,
		Path:    stripControls(path),
		Key:     stripControls(key),
		Line:    line,
		Column:  column,
		wrapped: wrapped,
	}
}

func (e *Error) Error() string {
	var builder strings.Builder

	switch e.Kind {
	case KindMissing:
		builder.WriteString("aws-cli-guard: policy: file missing")
	case KindRead:
		builder.WriteString("aws-cli-guard: policy: read failed")
	case KindPath:
		builder.WriteString("aws-cli-guard: policy: invalid path")
	case KindDecode:
		builder.WriteString("aws-cli-guard: policy: invalid document")
	case KindInvalid:
		builder.WriteString("aws-cli-guard: policy: invalid value")
	default:
		builder.WriteString("aws-cli-guard: policy: error")
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
