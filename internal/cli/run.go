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
func Run(args []string, stderr io.Writer) int {
	if len(args) < minArgsWithCommand {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	output.Error(stderr, usageLine)

	return output.CancelStatus
}
