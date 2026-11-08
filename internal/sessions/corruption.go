package sessions

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/files"
	"errors"
)

// Corruption reports whether err is a corrupt sessions file and builds the diagnosis note.
func Corruption(err error) (diagnose.Note, bool) {
	if err == nil {
		return emptyDiagnoseNote(), false
	}

	sessErr, ok := errors.AsType[*Error](err)
	if !ok {
		return emptyDiagnoseNote(), false
	}

	switch sessErr.Kind {
	case KindMissing, KindPath:
		return emptyDiagnoseNote(), false
	case KindInvalid:
		return noteFromSessionsError(sessErr, "invalid value"), true
	case KindRead:
		return noteFromSessionsError(sessErr, readFailurePhrase(sessErr)), true
	case KindDecode:
		decErr, decOK := errors.AsType[*decode.Error](sessErr.Unwrap())
		if !decOK || isProgrammerDecodeError(decErr) {
			return emptyDiagnoseNote(), false
		}

		return noteFromSessionsError(sessErr, decodeFailurePhrase(decErr)), true
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

func noteFromSessionsError(sessErr *Error, failure string) diagnose.Note {
	supported := append([]int(nil), sessErr.Supported...)

	return diagnose.Note{
		Class:      diagnose.ClassSessions(),
		Path:       sessErr.Path,
		Failure:    failure,
		Key:        sessErr.Key,
		Line:       sessErr.Line,
		Column:     sessErr.Column,
		Version:    sessErr.Version,
		HasVersion: sessErr.HasVersion,
		Supported:  supported,
	}
}

func readFailurePhrase(sessErr *Error) string {
	wrapped := sessErr.Unwrap()

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
