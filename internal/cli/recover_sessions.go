package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/sessions"
	"errors"
	"io"
	"time"
)

const (
	recoverTargetSessions     = "sessions"
	reasonRecoverSessionsMiss = "the sessions file does not exist"
	reasonRecoverSessionsWrt  = "the sessions file could not be written"
)

func recoverSessions(env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	err := sessions.Present(env)
	if err != nil {
		output.Error(stderr, recoverSessionsPresentReason(err))

		return output.CancelStatus
	}

	err = confirm(
		recoverTargetSessions,
		"type sessions to replace the sessions file with an empty record list",
	)
	if err != nil {
		output.Error(stderr, reasonRecoverConfirm)

		return output.CancelStatus
	}

	aside, err := sessions.Recover(env, time.Now())
	if err != nil {
		if aside != "" {
			output.Detail(stdout, "aside: "+aside)
		}

		output.Error(stderr, reasonRecoverSessionsWrt)

		return output.CancelStatus
	}

	output.Detail(stdout, "aside: "+aside)
	output.Detail(stdout, "wrote fresh sessions file")

	return 0
}

func recoverSessionsPresentReason(err error) string {
	if errors.Is(err, paths.ErrHomeUnset) {
		return reasonRecoverPath
	}

	if errors.Is(err, sessions.ErrMissing) {
		return reasonRecoverSessionsMiss
	}

	return reasonRecoverPath
}
