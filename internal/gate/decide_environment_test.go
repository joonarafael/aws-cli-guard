package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"testing"
)

func TestDecideEnvironmentDenyOverridesDefaultsAllowAndGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, `version = 2

[defaults]
action = "allow"

[environment]
action = "deny"

[credential_export]
action = "allow"
`)
	writeGateGrants(t, env, currentGrantBody(gateHandoffProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		append(env,
			"AWS_ACCESS_KEY_ID=AKIAEXAMPLE",
			"AWS_SECRET_ACCESS_KEY=secret",
		),
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "the policy denies this subject")
}

func TestDecideEnvironmentAllowOverridesDefaultsDeny(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, `version = 2

[defaults]
action = "deny"

[environment]
action = "allow"

[credential_export]
action = "allow"
`)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		append(env,
			"AWS_ACCESS_KEY_ID=AKIAEXAMPLE",
			"AWS_SECRET_ACCESS_KEY=secret",
		),
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, want true; stderr = %q", stderr.String())
	}
}
