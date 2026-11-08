package cli

import (
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/output"
	"io"
)

const (
	upgradeCommand      = "upgrade"
	reasonUpgradeUsage  = "usage: aws-guard upgrade [--prefix dir]"
	upgradeArgsNoPrefix = 2
	upgradeArgsPrefix   = 4
)

func runUpgrade(args []string, env []string, stdout, stderr io.Writer) int {
	prefix, ok := parseUpgradeArgs(args)
	if !ok {
		output.Error(stderr, reasonUpgradeUsage)

		return output.CancelStatus
	}

	_, err := install.Upgrade(install.Options{
		Env:           env,
		Prefix:        prefix,
		Executable:    "",
		BeforeUpgrade: func() error { return preflightUpgradeFiles(env) },
	}, stdout)
	if err != nil {
		reportInstallOrUpgradeFailure(stderr, env, err)

		return output.CancelStatus
	}

	return runMigrateAfterUpgrade(env, stdout, stderr, true)
}

func runInstallWithUpgrade(args []string, env []string, stdout, stderr io.Writer) int {
	prefix, ok := parseInstallArgs(args)
	if !ok {
		output.Error(stderr, reasonInstallUsage)

		return output.CancelStatus
	}

	result, err := install.Run(install.Options{
		Env:           env,
		Prefix:        prefix,
		Executable:    "",
		BeforeUpgrade: func() error { return preflightUpgradeFiles(env) },
	}, stdout)
	if err != nil {
		reportInstallOrUpgradeFailure(stderr, env, err)

		return output.CancelStatus
	}

	if result.Upgraded {
		return runMigrateAfterUpgrade(env, stdout, stderr, true)
	}

	return 0
}

func parseUpgradeArgs(args []string) (string, bool) {
	return parsePrefixArgs(args, upgradeArgsNoPrefix, upgradeArgsPrefix)
}

func parseInstallArgs(args []string) (string, bool) {
	return parsePrefixArgs(args, installArgsNoPrefix, installArgsPrefix)
}

func parsePrefixArgs(args []string, noPrefixLen, prefixLen int) (string, bool) {
	switch len(args) {
	case noPrefixLen:
		return "", true
	case prefixLen:
		if args[2] != flagPrefix || args[3] == "" {
			return "", false
		}

		return args[3], true
	default:
		return "", false
	}
}
