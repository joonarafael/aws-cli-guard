package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

const (
	progName       = "aws-guard"
	cliExtraArg    = "extra"
	flagPrefixTest = "--prefix"
)

func TestRunNoArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, stderr.String())
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, "unknown-cmd"}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, stderr.String())

	if strings.Contains(stderr.String(), "unknown-cmd") {
		t.Fatalf("stderr must not echo argv: %q", stderr.String())
	}
}

func assertUsageLine(t *testing.T, stderr string) {
	t.Helper()

	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	want := "aws-cli-guard: error: usage: aws-guard <command>"

	if len(lines) == 0 || lines[0] != want {
		t.Fatalf("stderr first line = %q want %q", lines[0], want)
	}
}
