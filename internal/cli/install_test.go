package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

const installSubcommand = "install"

func TestRunInstallUsageTooFewArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, installSubcommand, cliExtraArg},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertInstallUsage(t, stderr.String())
}

func TestRunInstallUsagePrefixForm(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, installSubcommand, "--prefix"},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertInstallUsage(t, stderr.String())
}

func TestRunInstallUsagePrefixEquals(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, installSubcommand, "--prefix=/tmp/x"},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertInstallUsage(t, stderr.String())
}

func assertInstallUsage(t *testing.T, stderr string) {
	t.Helper()

	line := strings.TrimSuffix(stderr, "\n")

	want := "aws-cli-guard: error: usage: aws-guard install [--prefix dir]"
	if line != want {
		t.Fatalf("stderr = %q want %q", line, want)
	}
}
