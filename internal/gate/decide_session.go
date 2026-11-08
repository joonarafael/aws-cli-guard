package gate

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/sessions"
	"errors"
	"io"
)

func sessionOffPermits(env []string, stderr io.Writer) (bool, bool) {
	sessFile, err := sessions.Load(env)
	if err != nil {
		cancelSessionsLoad(stderr, err)

		return false, false
	}

	if sessFile.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassSessions(), sessFile.Path, sessFile.Mode)
	}

	if session.PolicySkipped(sessFile, session.ID()) {
		return true, true
	}

	return false, true
}

func cancelSessionsLoad(stderr io.Writer, err error) {
	if note, ok := sessions.Corruption(err); ok {
		output.Cancel(stderr, reasonSessionsUnusable)
		diagnose.Write(stderr, note)

		return
	}

	var sessErr *sessions.Error

	if errors.As(err, &sessErr) && sessErr.Kind == sessions.KindPath {
		output.Cancel(stderr, reasonSessionsPathUnset)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Cancel(stderr, reasonSessionsPathUnset)

		return
	}

	output.Cancel(stderr, reasonSessionsReadFail)
}
