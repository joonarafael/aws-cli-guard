package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"strings"
	"testing"
)

const productMarkerPolicyPrefix = `version = 4

[defaults]
action = "allow"

[environment]
action = "allow"

[credential_export]
action = "allow"

[agents.detection]
environment_markers = true
`

func TestDecideProductMarkerDeniesRecognizedAgent(t *testing.T) {
	t.Parallel()

	markers := []string{
		"CLAUDE_CODE_CHILD_SESSION",
		"CURSOR_AGENT",
	}

	for _, marker := range markers {
		t.Run(marker, func(t *testing.T) {
			t.Parallel()

			env := gateEnv(t)
			env = append(env, marker+"=1")
			writeGatePolicy(t, env, productMarkerPolicyPrefix+`

[profiles.`+gateHandoffProfile+`]
action = "allow"

[agents.profiles.`+gateHandoffProfile+`]
action = "deny"
`)

			var stderr bytes.Buffer

			permitted := gate.Decide(
				[]string{
					gateProgName,
					gateProfileFlag,
					gateHandoffProfile,
					gateSTSService,
					gateGetCallerIdentityOp,
				},
				env,
				"",
				&stderr,
			)
			if permitted {
				t.Fatal("Decide = true, want false")
			}

			out := stderr.String()
			assertGateCancel(t, out)

			if !strings.Contains(out, "caller: recognized AI agent") {
				t.Fatalf("stderr = %q", out)
			}
		})
	}
}
