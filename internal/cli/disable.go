package cli

import (
	"aws-cli-guard/internal/output"
	"io"
)

const (
	disableCommand = "disable"
	enableCommand  = "enable"

	disableArgsExact = 2
	enableArgsExact  = 2

	reasonDisableUsage = "usage: aws-guard disable"
	reasonEnableUsage  = "usage: aws-guard enable"
)

func runDisable(args []string, _ []string, stdout, stderr io.Writer) int {
	if len(args) != disableArgsExact {
		output.Error(stderr, reasonDisableUsage)

		return output.CancelStatus
	}

	return runPersistentDisableScript(disableCommand, stdout, stderr)
}

func runEnable(args []string, _ []string, stdout, stderr io.Writer) int {
	if len(args) != enableArgsExact {
		output.Error(stderr, reasonEnableUsage)

		return output.CancelStatus
	}

	return runPersistentDisableScript(enableCommand, stdout, stderr)
}
