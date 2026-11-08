package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

const upgradeSubcommand = "upgrade"

func TestRunUpgradeUsageTooFewArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, upgradeSubcommand, cliExtraArg},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	line := strings.TrimSuffix(stderr.String(), "\n")

	want := "aws-cli-guard: error: usage: aws-guard upgrade [--prefix dir]"
	if line != want {
		t.Fatalf("stderr = %q want %q", line, want)
	}
}

func TestRunUpgradeUsagePrefixForm(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, upgradeSubcommand, "--prefix"},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	line := strings.TrimSuffix(stderr.String(), "\n")

	want := "aws-cli-guard: error: usage: aws-guard upgrade [--prefix dir]"
	if line != want {
		t.Fatalf("stderr = %q want %q", line, want)
	}
}
