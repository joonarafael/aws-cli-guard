package grants

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/files"
	"errors"
)

// Corruption reports whether err is a corrupt grants file and builds the diagnosis note.
func Corruption(err error) (diagnose.Note, bool) {
	if err == nil {
		return emptyDiagnoseNote(), false
	}

	grErr, ok := errors.AsType[*Error](err)
	if !ok {
		return emptyDiagnoseNote(), false
	}

	switch grErr.Kind {
	case KindPath:
		return emptyDiagnoseNote(), false
	case KindInvalid:
		return noteFromGrantsError(grErr, "invalid value"), true
	case KindRead:
		return noteFromGrantsError(grErr, readFailurePhrase(grErr)), true
	case KindDecode:
		decErr, decOK := errors.AsType[*decode.Error](grErr.Unwrap())
		if !decOK || isProgrammerDecodeError(decErr) {
			return emptyDiagnoseNote(), false
		}

		return noteFromGrantsError(grErr, decodeFailurePhrase(decErr)), true
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

func noteFromGrantsError(grErr *Error, failure string) diagnose.Note {
	supported := append([]int(nil), grErr.Supported...)

	return diagnose.Note{
		Class:      diagnose.ClassGrants(),
		Path:       grErr.Path,
		Failure:    failure,
		Key:        grErr.Key,
		Line:       grErr.Line,
		Column:     grErr.Column,
		Version:    grErr.Version,
		HasVersion: grErr.HasVersion,
		Supported:  supported,
	}
}

func readFailurePhrase(grErr *Error) string {
	wrapped := grErr.Unwrap()

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
