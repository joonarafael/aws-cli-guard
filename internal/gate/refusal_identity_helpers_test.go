package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"strings"
	"testing"
)

const (
	identityRoleARN          = "arn:aws:iam::123456789012:role/Admin"
	identityAccountID        = "123456789012"
	identitySSOAccountID     = "111111111111"
	identitySSORoleName      = "Developer"
	identityCredProcMarker   = "UNIQUE-CRED-PROC-REFUSAL-TEST"
	identitySessionStartURL  = "https://session.example/start"
	identityUnverifiedDetail = "identity: configured source only; the caller was not verified"
)

func denyArgs() []string {
	return []string{
		gateProgName,
		gateSTSService,
		gateGetCallerIdentityOp,
		gateProfileFlag,
		gateTestProfile,
	}
}

func decideDenied(t *testing.T, env []string, args []string) string {
	t.Helper()

	var stderr bytes.Buffer

	permitted := gate.Decide(args, env, "", &stderr)
	if permitted {
		t.Fatalf("Decide = true, want false; stderr = %q", stderr.String())
	}

	return stderr.String()
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Fatalf("stderr = %q, must not contain %q", haystack, needle)
	}
}

func assertDetailOrder(t *testing.T, stderr string, prefixes ...string) {
	t.Helper()

	lines := detailContentLines(stderr)
	if len(lines) < len(prefixes) {
		t.Fatalf("detail lines = %v, want at least %d", lines, len(prefixes))
	}

	for i, prefix := range prefixes {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Fatalf("detail[%d] = %q, want prefix %q; all = %v", i, lines[i], prefix, lines)
		}
	}
}

func detailContentLines(stderr string) []string {
	var out []string

	for line := range strings.SplitSeq(stderr, "\n") {
		if !strings.HasPrefix(line, "aws-cli-guard: ") {
			continue
		}

		if strings.HasPrefix(line, "aws-cli-guard: error:") ||
			strings.HasPrefix(line, "aws-cli-guard: warning:") {
			continue
		}

		out = append(out, strings.TrimPrefix(line, "aws-cli-guard: "))
	}

	return out
}
