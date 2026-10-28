package policy

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/files"
	"errors"
)

// Corruption reports whether err is a corrupt policy file and builds the diagnosis note.
func Corruption(err error) (diagnose.Note, bool) {
	if err == nil {
		return emptyDiagnoseNote(), false
	}

	polErr, ok := errors.AsType[*Error](err)
	if !ok {
		return emptyDiagnoseNote(), false
	}

	switch polErr.Kind {
	case KindMissing, KindPath:
		return emptyDiagnoseNote(), false
	case KindInvalid:
		return noteFromPolicyError(polErr, "invalid value"), true
	case KindRead:
		return noteFromPolicyError(polErr, readFailurePhrase(polErr)), true
	case KindDecode:
		decErr, decOK := errors.AsType[*decode.Error](polErr.Unwrap())
		if !decOK || isProgrammerDecodeError(decErr) {
			return emptyDiagnoseNote(), false
		}

		return noteFromPolicyError(polErr, decodeFailurePhrase(decErr)), true
	default:
		return emptyDiagnoseNote(), false
	}
}

func emptyDiagnoseNote() diagnose.Note {
	var note diagnose.Note

	return note
}

func noteFromPolicyError(polErr *Error, failure string) diagnose.Note {
	supported := append([]int(nil), polErr.Supported...)

	return diagnose.Note{
		Class:      diagnose.ClassPolicy(),
		Path:       polErr.Path,
		Failure:    failure,
		Key:        polErr.Key,
		Line:       polErr.Line,
		Column:     polErr.Column,
		Version:    polErr.Version,
		HasVersion: polErr.HasVersion,
		Supported:  supported,
	}
}

func readFailurePhrase(polErr *Error) string {
	wrapped := polErr.Unwrap()

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
