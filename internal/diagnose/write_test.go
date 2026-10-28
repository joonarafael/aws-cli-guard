package diagnose_test

import (
	"aws-cli-guard/internal/diagnose"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestWriteLooseModeWarnsOn0644(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.ClassPolicy(), "/cfg/config.toml", 0o644)

	text := strings.TrimSuffix(buf.String(), "\n")

	want := `aws-cli-guard: warning: policy file "/cfg/config.toml" has loose permissions (0644)`
	if text != want {
		t.Fatalf("got %q want %q", text, want)
	}
}

func TestWriteLooseModeSilentOn0600(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.ClassPolicy(), "/cfg/config.toml", 0o600)

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %q", buf.String())
	}
}

func TestWriteLooseModeSilentOn0400(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.ClassPolicy(), "/cfg/config.toml", 0o400)

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %q", buf.String())
	}
}

func TestWriteIncludesRequiredLines(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.Write(&buf, diagnose.Note{
		Class:      diagnose.ClassPolicy(),
		Path:       "/cfg/config.toml",
		Failure:    "unsupported version",
		Key:        "",
		Line:       0,
		Column:     0,
		Version:    99,
		HasVersion: true,
		Supported:  []int{1},
	})

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	want := []string{
		"aws-cli-guard: file: policy",
		"aws-cli-guard: path: /cfg/config.toml",
		"aws-cli-guard: failure: unsupported version",
		"aws-cli-guard: version: 99",
		"aws-cli-guard: supported versions: 1",
		"aws-cli-guard: the file was not modified",
		"aws-cli-guard: fix the file in place, or replace it with a fresh one",
		"aws-cli-guard: aws-guard recover policy",
	}

	if len(lines) != len(want) {
		t.Fatalf("line count = %d, want %d:\n%q", len(lines), len(want), buf.String())
	}

	for i, line := range want {
		if lines[i] != line {
			t.Fatalf("line %d: got %q want %q", i, lines[i], line)
		}
	}
}

func TestWriteUnknownClassOmitsRecover(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.Write(&buf, diagnose.Note{
		Class:      diagnose.Class{Name: "", Recover: ""},
		Path:       "/x",
		Failure:    "unreadable",
		Key:        "",
		Line:       0,
		Column:     0,
		Version:    0,
		HasVersion: false,
		Supported:  nil,
	})

	text := buf.String()
	if strings.Contains(text, "recover") {
		t.Fatalf("unexpected recover line: %q", text)
	}

	if !strings.Contains(text, "file: unknown") {
		t.Fatalf("expected unknown class: %q", text)
	}
}

func TestWriteLooseModeUsesFileLabelForEmptyClass(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.Class{Name: "", Recover: ""}, "/x", os.FileMode(0o644))

	text := strings.TrimSuffix(buf.String(), "\n")
	if !strings.Contains(text, `file "`) {
		t.Fatalf("expected generic file label: %q", text)
	}
}
