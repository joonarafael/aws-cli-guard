package output_test

import (
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

func TestCancelWritesRequiredLines(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	output.Cancel(&buf, "the real AWS CLI was not found")

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}

	if lines[0] != "aws-cli-guard: error: command was not executed" {
		t.Fatalf("line 0: %q", lines[0])
	}

	want := "aws-cli-guard: error: AWS CLI Guard stopped it: the real AWS CLI was not found"
	if lines[1] != want {
		t.Fatalf("line 1: %q want %q", lines[1], want)
	}
}

func TestCancelStripsControlCharacters(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	reason := "before\nafter\rafter2\x00\x1bmore"
	output.Cancel(&buf, reason)

	text := buf.String()
	if strings.Contains(text, "\n\n") {
		t.Fatalf("expected single physical lines, got %q", text)
	}

	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	wantReason := "aws-cli-guard: error: AWS CLI Guard stopped it: beforeafterafter2more"
	if lines[1] != wantReason {
		t.Fatalf("sanitized reason line: %q want %q", lines[1], wantReason)
	}
}

func TestWarningPrefixAndStripping(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	output.Warning(&buf, "old\x0acli")

	line := strings.TrimSuffix(buf.String(), "\n")

	want := "aws-cli-guard: warning: oldcli"
	if line != want {
		t.Fatalf("got %q want %q", line, want)
	}
}

func TestCancelStatusConstant(t *testing.T) {
	t.Parallel()

	if output.CancelStatus != 3 {
		t.Fatalf("CancelStatus = %d, want 3", output.CancelStatus)
	}
}

func TestErrorPrefixAndStripping(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	output.Error(&buf, "usage\x0aline")

	line := strings.TrimSuffix(buf.String(), "\n")

	want := "aws-cli-guard: error: usageline"
	if line != want {
		t.Fatalf("got %q want %q", line, want)
	}
}

func TestDetailPrefixAndStripping(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	output.Detail(&buf, "path\x0a/home")

	line := strings.TrimSuffix(buf.String(), "\n")

	want := "aws-cli-guard: path/home"
	if line != want {
		t.Fatalf("got %q want %q", line, want)
	}
}
