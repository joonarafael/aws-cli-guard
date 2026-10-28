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

	var buf bytes.Buffer

	status := cli.Run([]string{progName}, &buf)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, buf.String())
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	status := cli.Run([]string{progName, "init"}, &buf)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, buf.String())

	if strings.Contains(buf.String(), "init") {
		t.Fatalf("stderr must not echo argv: %q", buf.String())
	}
}

func TestRunFlagLikeCommand(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	status := cli.Run([]string{progName, "--help"}, &buf)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	assertUsageLine(t, buf.String())

	if strings.Contains(buf.String(), "--help") {
		t.Fatalf("stderr must not echo argv: %q", buf.String())
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
