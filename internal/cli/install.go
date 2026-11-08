package cli

import "io"

const (
	installCommand      = "install"
	installArgsNoPrefix = 2
	installArgsPrefix   = 4
	flagPrefix          = "--prefix"

	reasonInstallUsage = "usage: aws-guard install [--prefix dir]"
)

func runInstall(args []string, env []string, stdout, stderr io.Writer) int {
	return runInstallWithUpgrade(args, env, stdout, stderr)
}
