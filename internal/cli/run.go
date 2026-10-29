package cli

import (
	"aws-cli-guard/internal/output"
	"io"
)

const (
	usageLine          = "usage: aws-guard <command>"
	minArgsWithCommand = 2
)

// Run dispatches aws-guard subcommands. args[0] is the program name and is ignored.
func Run(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) < minArgsWithCommand {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	switch args[1] {
	case initCommand:
		return runInit(args, env, stdout, stderr)
	default:
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}
}
