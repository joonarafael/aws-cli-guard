package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/version"
	"bytes"
	"strings"
	"testing"
)

func TestRunHelpOnStdout(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"help", "--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			status := cli.Run([]string{progName, arg}, nil, &stdout, &stderr)
			if status != 0 {
				t.Fatalf("status = %d, want 0", status)
			}

			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}

			text := stdout.String()
			if !strings.Contains(text, "aws-guard commands:") {
				t.Fatalf("stdout missing header: %q", text)
			}

			if !strings.Contains(text, "allow <profile>") {
				t.Fatalf("stdout missing allow synopsis: %q", text)
			}
		})
	}
}

func TestRunVersionOnStdout(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"version", "--version"} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			status := cli.Run([]string{progName, arg}, nil, &stdout, &stderr)
			if status != 0 {
				t.Fatalf("status = %d, want 0", status)
			}

			want := "aws-guard " + version.Version + "\n"
			if stdout.String() != want {
				t.Fatalf("stdout = %q, want %q", stdout.String(), want)
			}
		})
	}
}

func TestRunNoArgsListsCommandsOnStderr(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	text := stderr.String()
	assertUsageLine(t, text)

	if !strings.Contains(text, "aws-guard commands:") {
		t.Fatalf("stderr missing command list: %q", text)
	}
}

func TestRunUnknownCommandListsCommandsOnStderr(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, "unknown-cmd"}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	text := stderr.String()
	assertUsageLine(t, text)

	if !strings.Contains(text, "aws-guard commands:") {
		t.Fatalf("stderr missing command list: %q", text)
	}

	if strings.Contains(text, "unknown-cmd") {
		t.Fatalf("stderr must not echo argv: %q", text)
	}
}

func TestHelpTableCoversDispatch(t *testing.T) {
	t.Parallel()

	covered := make(map[string]bool)

	for _, name := range cli.HelpTableCommandNames() {
		covered[name] = true
	}

	for _, name := range cli.DispatchCommandNames() {
		if !covered[name] {
			t.Fatalf("dispatch command %q missing from help table", name)
		}
	}
}
