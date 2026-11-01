package gate_test

import (
	"aws-cli-guard/internal/output"
	"strings"
	"testing"
)

func TestDecideCancelStatusConstant(t *testing.T) {
	t.Parallel()

	if output.CancelStatus != 3 {
		t.Fatalf("CancelStatus = %d, want 3", output.CancelStatus)
	}
}

func assertGateCancel(t *testing.T, stderr string) {
	t.Helper()

	if !strings.Contains(stderr, "aws-cli-guard: error: command was not executed") {
		t.Fatalf("missing cancellation line: %q", stderr)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Fatalf("stderr = %q, want substring %q", haystack, needle)
	}
}

func assertNoSecret(t *testing.T, stderr string) {
	t.Helper()

	if strings.Contains(stderr, gateSecretArg) {
		t.Fatalf("stderr leaked argv token")
	}
}
