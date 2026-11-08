package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/sessions"
	"io"
	"time"
)

const (
	offCommand = "off"
	onCommand  = "on"

	offArgsExact = 2
	onArgsExact  = 2

	reasonOffUsage          = "usage: aws-guard off"
	reasonOnUsage           = "usage: aws-guard on"
	reasonSessionNotConfirm = "the session change was not confirmed"
	reasonSessionNoTTY      = "this process has no controlling terminal"
	reasonSessionWrite      = "the session file could not be written"

	offConfirmPrompt = "type off to turn policy off for this terminal session"

	offSuccessLine1 = "this terminal session is off and policy is not applied"
	offSuccessLine2 = "a new terminal starts with policy on"
	offSuccessLine3 = "aws-guard on turns policy back on for this session"
	onSuccessLine   = "this terminal session is on and policy applies"
)

type ttyCaptureFunc func() (sessions.TTY, error)

func runOff(
	args []string,
	env []string,
	stdout, stderr io.Writer,
	confirm confirmFunc,
	capture ttyCaptureFunc,
) int {
	if len(args) != offArgsExact {
		output.Error(stderr, reasonOffUsage)

		return output.CancelStatus
	}

	err := confirm(offCommand, offConfirmPrompt)
	if err != nil {
		output.Error(stderr, reasonSessionNotConfirm)

		return output.CancelStatus
	}

	tty, err := capture()
	if err != nil {
		output.Error(stderr, reasonSessionNoTTY)

		return output.CancelStatus
	}

	err = sessions.Append(env, sessions.AppendRequest{
		SessionID: session.ID(),
		Effect:    sessions.EffectOff,
		TTY:       &tty,
	}, time.Now())
	if err != nil {
		output.Error(stderr, reasonSessionWrite)

		return output.CancelStatus
	}

	output.Detail(stdout, offSuccessLine1)
	output.Detail(stdout, offSuccessLine2)
	output.Detail(stdout, offSuccessLine3)

	return 0
}

func runOn(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != onArgsExact {
		output.Error(stderr, reasonOnUsage)

		return output.CancelStatus
	}

	var tty *sessions.TTY

	if current, ok := session.CurrentControllingTTY(); ok {
		tty = &current
	}

	err := sessions.Append(env, sessions.AppendRequest{
		SessionID: session.ID(),
		Effect:    sessions.EffectOn,
		TTY:       tty,
	}, time.Now())
	if err != nil {
		output.Error(stderr, reasonSessionWrite)

		return output.CancelStatus
	}

	output.Detail(stdout, onSuccessLine)

	return 0
}
