package decode

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Programmer-error sentinels. FS-3 must not treat these as corrupt files.
var (
	ErrFormat = errors.New("aws-cli-guard: decode: invalid format")
	ErrDest   = errors.New(
		"aws-cli-guard: decode: dest must be a non-nil pointer to struct",
	)
	ErrVersionField = errors.New(
		"aws-cli-guard: decode: dest struct lacks version int field with matching tag",
	)
	ErrSupported          = errors.New("aws-cli-guard: decode: supported version list is empty")
	ErrSyntax             = errors.New("aws-cli-guard: decode: invalid syntax")
	ErrDuplicateKey       = errors.New("aws-cli-guard: decode: duplicate key")
	ErrMissingVersion     = errors.New("aws-cli-guard: decode: missing version")
	ErrVersionType        = errors.New("aws-cli-guard: decode: version is not an integer")
	ErrUnsupportedVersion = errors.New("aws-cli-guard: decode: unsupported version")
	ErrUnknownField       = errors.New("aws-cli-guard: decode: unknown field")
	ErrType               = errors.New("aws-cli-guard: decode: invalid value type")
)

// Kind classifies a document decode failure.
type Kind int

// Document failure kinds returned in decode.Error.
const (
	KindSyntax Kind = iota + 1
	KindDuplicateKey
	KindMissingVersion
	KindVersionType
	KindUnsupportedVersion
	KindUnknownField
	KindType
)

// Error is a secret-safe document decode failure.
type Error struct {
	Kind       Kind
	Key        string
	Line       int
	Column     int
	Version    int
	HasVersion bool
	Supported  []int
	wrapped    error
}

func newDocError(
	kind Kind,
	wrapped error,
	key string,
	line, column int,
	version int,
	hasVersion bool,
	supported []int,
) *Error {
	suppCopy := append([]int(nil), supported...)

	return &Error{
		Kind:       kind,
		Key:        stripControls(key),
		Line:       line,
		Column:     column,
		Version:    version,
		HasVersion: hasVersion,
		Supported:  suppCopy,
		wrapped:    wrapped,
	}
}

func (e *Error) Error() string {
	var builder strings.Builder

	switch e.Kind {
	case KindSyntax:
		builder.WriteString("aws-cli-guard: decode: invalid syntax")
	case KindDuplicateKey:
		builder.WriteString("aws-cli-guard: decode: duplicate key")

		if e.Key != "" {
			fmt.Fprintf(&builder, " %q", e.Key)
		}
	case KindMissingVersion:
		builder.WriteString("aws-cli-guard: decode: missing required version field")
	case KindVersionType:
		builder.WriteString("aws-cli-guard: decode: version must be an integer")
	case KindUnsupportedVersion:
		builder.WriteString("aws-cli-guard: decode: unsupported file version")

		if e.HasVersion {
			fmt.Fprintf(&builder, " %d", e.Version)
		}

		if len(e.Supported) > 0 {
			fmt.Fprintf(&builder, " (supported: %s)", formatIntList(e.Supported))
		}
	case KindUnknownField:
		builder.WriteString("aws-cli-guard: decode: unknown field")

		if e.Key != "" {
			fmt.Fprintf(&builder, " %q", e.Key)
		}
	case KindType:
		builder.WriteString("aws-cli-guard: decode: invalid value type")

		if e.Key != "" {
			fmt.Fprintf(&builder, " for field %q", e.Key)
		}
	default:
		builder.WriteString("aws-cli-guard: decode: invalid document")
	}

	if e.Line > 0 {
		fmt.Fprintf(&builder, " at line %d column %d", e.Line, e.Column)
	}

	return builder.String()
}

func (e *Error) Unwrap() error {
	return e.wrapped
}

func formatIntList(vals []int) string {
	parts := make([]string, len(vals))
	for i, val := range vals {
		parts[i] = strconv.Itoa(val)
	}

	return strings.Join(parts, ", ")
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

func joinKey(parts []string) string {
	return strings.Join(parts, ".")
}
