package verbose_test

import (
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/verbose"
	"bytes"
	"strings"
	"testing"
)

func TestWriteTightEmitsNothing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	event := verbose.Event{Outcome: "allow"} //nolint:exhaustruct_v5 // tight mode ignores fields.
	verbose.Write(&buf, settings.LevelTight, event)

	if buf.Len() != 0 {
		t.Fatalf("buf = %q", buf.String())
	}
}

func TestWriteVerboseOrderAndSkipEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	verbose.Write(
		&buf,
		settings.LevelVerbose,
		verbose.Event{ //nolint:exhaustruct_v5 // partial block for order test.
			Subject: "named profile prod",
			Rule:    "profile allow",
			AWS:     "/usr/bin/aws",
			Outcome: "allow",
		},
	)

	out := buf.String()

	want := []string{
		"aws-cli-guard: subject: named profile prod",
		"aws-cli-guard: rule: profile allow",
		"aws-cli-guard: aws: /usr/bin/aws",
		"aws-cli-guard: outcome: allow",
	}

	idx := 0

	for _, line := range want {
		pos := strings.Index(out, line)
		if pos < 0 {
			t.Fatalf("missing %q in %q", line, out)
		}

		if pos < idx {
			t.Fatalf("line %q out of order in %q", line, out)
		}

		idx = pos
	}

	if strings.Contains(out, "grant:") {
		t.Fatal("unexpected grant line")
	}
}
