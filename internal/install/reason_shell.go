package install

import (
	"aws-cli-guard/internal/files"
	"errors"
)

func shellReason(err error) (string, bool) {
	switch {
	case errors.Is(err, ErrShellStartupTooLarge):
		return reasonShellStartupTooLarge, true
	case errors.Is(err, ErrShellMarkersConflict):
		return reasonShellMarkersConflict, true
	case errors.Is(err, ErrShellUnsupportedLineEndings):
		return reasonShellLineEndings, true
	case errors.Is(err, ErrShellStartupUpdate):
		return reasonShellUpdateFailed, true
	case errors.Is(err, ErrShellNameUnsupported):
		return reasonShellNameUnsupported, true
	case errors.Is(err, files.ErrSymlink):
		return reasonShellStartupSymlink, true
	case errors.Is(err, files.ErrNotRegular):
		return reasonShellStartupNotReg, true
	default:
		return "", false
	}
}
