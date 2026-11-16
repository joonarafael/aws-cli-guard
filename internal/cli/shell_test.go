package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

const shellSubcommand = "shell"

func TestRunShellUsageTooFewArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, shellSubcommand}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertShellUsage(t, stderr.String())
}

func TestRunShellUsageUnknownShell(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, shellSubcommand, "fish"},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertShellUsage(t, stderr.String())
}

func TestRunShellUsagePrefixWithoutDir(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, shellSubcommand, "bash", flagPrefixTest},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertShellUsage(t, stderr.String())
}

func TestRunShellUsageRemoveWithExtraArg(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run(
		[]string{progName, shellSubcommand, "bash", "--remove", cliExtraArg},
		nil,
		&bytes.Buffer{},
		&stderr,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertShellUsage(t, stderr.String())
}

func assertShellUsage(t *testing.T, stderr string) {
	t.Helper()

	want := "aws-cli-guard: error: usage: aws-guard shell bash|zsh [--prefix dir] [--remove]"
	if strings.TrimSuffix(stderr, "\n") != want {
		t.Fatalf("stderr = %q, want %q", stderr, want)
	}
}
