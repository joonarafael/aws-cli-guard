package settings

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/files"
	"errors"
)

// Corruption reports whether err is a corrupt settings file and builds the diagnosis note.
func Corruption(err error) (diagnose.Note, bool) {
	if err == nil {
		return emptyDiagnoseNote(), false
	}

	setErr, ok := errors.AsType[*Error](err)
	if !ok {
		return emptyDiagnoseNote(), false
	}

	switch setErr.Kind {
	case KindMissing, KindPath:
		return emptyDiagnoseNote(), false
	case KindInvalid:
		return noteFromSettingsError(setErr, "invalid value"), true
	case KindRead:
		return noteFromSettingsError(setErr, readFailurePhrase(setErr)), true
	case KindDecode:
		decErr, decOK := errors.AsType[*decode.Error](setErr.Unwrap())
		if !decOK || isProgrammerDecodeError(decErr) {
			return emptyDiagnoseNote(), false
		}

		return noteFromSettingsError(setErr, decodeFailurePhrase(decErr)), true
	case KindWrite:
		return emptyDiagnoseNote(), false
	default:
		return emptyDiagnoseNote(), false
	}
}

func emptyDiagnoseNote() diagnose.Note {
	var note diagnose.Note

	return note
}

func noteFromSettingsError(setErr *Error, failure string) diagnose.Note {
	supported := append([]int(nil), setErr.Supported...)

	return diagnose.Note{
		Class:      diagnose.ClassSettings(),
		Path:       setErr.Path,
		Failure:    failure,
		Key:        setErr.Key,
		Line:       setErr.Line,
		Column:     setErr.Column,
		Version:    setErr.Version,
		HasVersion: setErr.HasVersion,
		Supported:  supported,
	}
}

func readFailurePhrase(setErr *Error) string {
	wrapped := setErr.Unwrap()

	switch {
	case errors.Is(wrapped, ErrTooLarge):
		return "file too large"
	case errors.Is(wrapped, files.ErrSymlink):
		return "symlink"
	case errors.Is(wrapped, files.ErrNotRegular):
		return "not a regular file"
	default:
		return "unreadable"
	}
}

func decodeFailurePhrase(decErr *decode.Error) string {
	switch decErr.Kind {
	case decode.KindSyntax:
		return "invalid syntax"
	case decode.KindDuplicateKey:
		return "duplicate key"
	case decode.KindMissingVersion:
		return "missing version"
	case decode.KindVersionType:
		return "version is not an integer"
	case decode.KindUnsupportedVersion:
		return "unsupported version"
	case decode.KindUnknownField:
		return "unknown field"
	case decode.KindType:
		return "invalid value type"
	default:
		return "invalid syntax"
	}
}

func isProgrammerDecodeError(decErr *decode.Error) bool {
	wrapped := decErr.Unwrap()

	return errors.Is(wrapped, decode.ErrFormat) ||
		errors.Is(wrapped, decode.ErrDest) ||
		errors.Is(wrapped, decode.ErrVersionField) ||
		errors.Is(wrapped, decode.ErrSupported)
}
