package session_test

import (
	"aws-cli-guard/internal/session"
	"strings"
	"testing"
)

func TestIDFormat(t *testing.T) {
	t.Parallel()

	got := session.ID()
	if got == "" {
		t.Fatal("empty session id")
	}

	if !strings.HasPrefix(got, "sid:") && !strings.HasPrefix(got, "ppid:") {
		t.Fatalf("unexpected format: %q", got)
	}
}
