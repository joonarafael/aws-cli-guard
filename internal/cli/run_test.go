package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

const progName = "aws-guard"

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

	status := cli.Run([]string{progName, "allow"}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, stderr.String())

	if strings.Contains(stderr.String(), "allow") {
		t.Fatalf("stderr must not echo argv: %q", stderr.String())
	}
}

func TestRunFlagLikeCommand(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, "--help"}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, stderr.String())

	if strings.Contains(stderr.String(), "--help") {
		t.Fatalf("stderr must not echo argv: %q", stderr.String())
	}
}

func assertUsageLine(t *testing.T, stderr string) {
	t.Helper()

	line := strings.TrimSuffix(stderr, "\n")

	want := "aws-cli-guard: error: usage: aws-guard <command>"
	if line != want {
		t.Fatalf("stderr = %q want %q", line, want)
	}
}
