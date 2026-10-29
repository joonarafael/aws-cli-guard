package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/terminal"
	"io"
)

const (
	usageLine          = "usage: aws-guard <command>"
	minArgsWithCommand = 2
)

// Run dispatches aws-guard subcommands. args[0] is the program name and is ignored.
func Run(args []string, env []string, stdout, stderr io.Writer) int {
	return Dispatch(args, env, stdout, stderr, terminal.Confirm)
}

// Dispatch runs aws-guard subcommands with the given recover confirmation hook.
func Dispatch(args []string, env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	if len(args) < minArgsWithCommand {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	switch args[1] {
	case initCommand:
		return runInit(args, env, stdout, stderr)
	case recoverCommand:
		return runRecover(args, env, stdout, stderr, confirm)
	default:
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}
}
