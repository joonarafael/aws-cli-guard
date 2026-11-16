//nolint:testpackage // unit tests for unexported shell marker editing.
package install

import (
	"errors"
	"strings"
	"testing"
)

func TestEditShellStartupAppend(t *testing.T) {
	t.Parallel()

	const prefix = "/opt/aws-cli-guard"

	got, changed, err := editShellStartup("", prefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if !changed {
		t.Fatal("want changed")
	}

	if !strings.Contains(got, "export PATH='"+prefix+":$PATH'") {
		t.Fatalf("got %q", got)
	}
}

func TestEditShellStartupIdempotent(t *testing.T) {
	t.Parallel()

	const prefix = "/opt/aws-cli-guard"

	before, _, err := editShellStartup("", prefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	after, changed, err := editShellStartup(before, prefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if changed {
		t.Fatalf("got %q, want unchanged", after)
	}
}

func TestEditShellStartupReplacePrefix(t *testing.T) {
	t.Parallel()

	const (
		oldPrefix = "/old"
		newPrefix = "/new"
	)

	before, _, err := editShellStartup("existing\n", oldPrefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	after, changed, err := editShellStartup(before, newPrefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if !changed {
		t.Fatal("want changed")
	}

	if !strings.Contains(after, newPrefix) || strings.Contains(after, oldPrefix) {
		t.Fatalf("got %q", after)
	}

	if !strings.HasPrefix(after, "existing\n") {
		t.Fatalf("got %q", after)
	}
}

func TestEditShellStartupRemove(t *testing.T) {
	t.Parallel()

	const prefix = "/opt/aws-cli-guard"

	before, _, err := editShellStartup("keep\n", prefix, false)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	after, changed, err := editShellStartup(before, prefix, true)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if !changed {
		t.Fatal("want changed")
	}

	if after != "keep\n" {
		t.Fatalf("got %q", after)
	}
}

func TestEditShellStartupRemoveMissingBlock(t *testing.T) {
	t.Parallel()

	const content = "plain\n"

	after, changed, err := editShellStartup(content, "/opt", true)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	if changed || after != content {
		t.Fatalf("got %q changed=%v", after, changed)
	}
}

func TestEditShellStartupConflictingMarkers(t *testing.T) {
	t.Parallel()

	content := "# aws-cli-guard begin\n# aws-cli-guard begin\n# aws-cli-guard end\n"

	_, _, err := editShellStartup(content, "/opt", false)
	if !errors.Is(err, ErrShellMarkersConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestEditShellStartupCarriageReturn(t *testing.T) {
	t.Parallel()

	_, _, err := editShellStartup("foo\r\n", "/opt", false)
	if !errors.Is(err, ErrShellUnsupportedLineEndings) {
		t.Fatalf("err = %v", err)
	}
}
