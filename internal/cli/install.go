package cli

import (
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/output"
	"io"
)

const (
	installCommand      = "install"
	installArgsNoPrefix = 2
	installArgsPrefix   = 4
	flagPrefix          = "--prefix"

	reasonInstallUsage = "usage: aws-guard install [--prefix dir]"
)

func runInstall(args []string, env []string, stdout, stderr io.Writer) int {
	prefix, ok := parseInstallArgs(args)
	if !ok {
		output.Error(stderr, reasonInstallUsage)

		return output.CancelStatus
	}

	err := install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: "",
	}, stdout)
	if err != nil {
		output.Error(stderr, install.Reason(err))

		return output.CancelStatus
	}

	return 0
}

func parseInstallArgs(args []string) (string, bool) {
	switch len(args) {
	case installArgsNoPrefix:
		return "", true
	case installArgsPrefix:
		if args[2] != flagPrefix || args[3] == "" {
			return "", false
		}

		return args[3], true
	default:
		return "", false
	}
}
