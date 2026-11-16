package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/version"
	"fmt"
	"io"
	"strings"
)

const (
	helpCommand    = "help"
	versionCommand = "version"
	usagePrefix    = "usage: aws-guard "
)

type helpEntry struct {
	command     string
	synopsis    string
	description string
}

func dispatchCommandNames() []string {
	return []string{
		initCommand,
		allowCommand,
		revokeCommand,
		policyCommand,
		configCommand,
		grantsCommand,
		recoverCommand,
		migrateCommand,
		installCommand,
		upgradeCommand,
		offCommand,
		onCommand,
		disableCommand,
		enableCommand,
		shellCommand,
		helpCommand,
		versionCommand,
	}
}

// HelpTableCommandNames lists primary command names shown in the help table.
func HelpTableCommandNames() []string {
	rows := helpEntries()
	names := make([]string, len(rows))

	for i, row := range rows {
		names[i] = row.command
	}

	return names
}

func helpEntries() []helpEntry {
	return []helpEntry{
		{initCommand, "init", "Create default policy and settings files"},
		{allowCommand, usageSynopsis(reasonAllowUsage), "Grant temporary access to a profile"},
		{
			revokeCommand,
			usageSynopsis(reasonRevokeUsage),
			"Remove grants for a profile or all profiles",
		},
		{
			policyCommand,
			usageSynopsis(reasonPolicyUsage),
			"List profiles or explain policy for one profile",
		},
		{configCommand, usageSynopsis(reasonConfigUsage), "Validate policy and settings files"},
		{grantsCommand, usageSynopsis(reasonGrantsUsage), "List active and expired grants"},
		{
			recoverCommand,
			usageSynopsis(reasonRecoverUsage),
			"Restore policy, grants, settings, cache, or sessions",
		},
		{
			migrateCommand,
			usageSynopsis(reasonMigrateUsage),
			"Upgrade policy and grants on disk to the current schema",
		},
		{
			installCommand,
			usageSynopsis(reasonInstallUsage),
			"Install binaries into a prefix from release assets",
		},
		{upgradeCommand, usageSynopsis(reasonUpgradeUsage), "Upgrade an existing install prefix"},
		{offCommand, usageSynopsis(reasonOffUsage), "Turn policy off for this terminal session"},
		{onCommand, usageSynopsis(reasonOnUsage), "Turn policy back on for this terminal session"},
		{disableCommand, usageSynopsis(reasonDisableUsage), "Swap the wrapper aside until enable"},
		{enableCommand, usageSynopsis(reasonEnableUsage), "Restore the wrapper after disable"},
		{
			shellCommand,
			usageSynopsis(reasonShellUsage),
			"Add or remove shell startup PATH integration",
		},
		{helpCommand, "help", "List aws-guard commands"},
		{versionCommand, "version", "Print the aws-guard version"},
	}
}

// DispatchCommandNames lists top-level commands handled by Dispatch.
func DispatchCommandNames() []string {
	names := dispatchCommandNames()
	out := make([]string, len(names))

	copy(out, names)

	return out
}

func usageSynopsis(usageLine string) string {
	return strings.TrimPrefix(usageLine, usagePrefix)
}

func printCommandList(w io.Writer) {
	for _, row := range helpEntries() {
		_, _ = fmt.Fprintf(w, "  %s\n    %s\n", row.synopsis, row.description)
	}
}

func runHelp(stdout io.Writer) int {
	_, _ = fmt.Fprintln(stdout, "aws-guard commands:")
	printCommandList(stdout)

	return 0
}

func runVersion(stdout io.Writer) int {
	_, _ = fmt.Fprintf(stdout, "aws-guard %s\n", version.Version)

	return 0
}

func printUsageWithCommands(stderr io.Writer) {
	output.Error(stderr, usageLine)
	_, _ = fmt.Fprintln(stderr, "aws-guard commands:")
	printCommandList(stderr)
}
