package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/terminal"
	"io"
)

const (
	usageLine          = "usage: aws-guard <command>"
	minArgsWithCommand = 2
)

// Run dispatches aws-guard subcommands. args[0] is the program name and is ignored.
func Run(args []string, env []string, stdout, stderr io.Writer) int {
	return Dispatch(args, env, stdout, stderr, terminal.Confirm, terminal.ReadLine)
}

// Dispatch runs aws-guard subcommands with the given recover confirmation hook.
// When readLine is omitted, init runs headless for updates.
func Dispatch(
	args []string,
	env []string,
	stdout, stderr io.Writer,
	confirm confirmFunc,
	readLine ...readLineFunc,
) int {
	updatesReadLine := headlessUpdatesReadLine
	if len(readLine) > 0 {
		updatesReadLine = readLine[0]
	}

	if len(args) < minArgsWithCommand {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	switch args[1] {
	case initCommand:
		return runInit(args, env, stdout, stderr, updatesReadLine)
	case recoverCommand:
		return runRecover(args, env, stdout, stderr, confirm)
	case allowCommand:
		return runAllow(args, env, stdout, stderr, confirm)
	case revokeCommand:
		return runRevoke(args, env, stdout, stderr)
	case policyCommand:
		return runPolicy(args, env, stdout, stderr)
	case configCommand:
		return runConfig(args, env, stdout, stderr)
	case grantsCommand:
		return runGrants(args, env, stdout, stderr)
	case migrateCommand:
		return runMigrate(args, env, stdout, stderr)
	case installCommand:
		return runInstall(args, env, stdout, stderr)
	case upgradeCommand:
		return runUpgrade(args, env, stdout, stderr)
	case offCommand:
		return runOff(args, env, stdout, stderr, confirm, session.CaptureControllingTTY)
	case onCommand:
		return runOn(args, env, stdout, stderr)
	case disableCommand:
		return runDisable(args, env, stdout, stderr)
	case enableCommand:
		return runEnable(args, env, stdout, stderr)
	default:
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}
}
