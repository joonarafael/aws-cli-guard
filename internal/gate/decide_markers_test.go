package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"strings"
	"testing"
)

const v3MarkersPolicyPrefix = `version = 3

[defaults]
action = "allow"

[environment]
action = "allow"

[credential_export]
action = "allow"
`

func TestDecideVersion3EnvironmentMarkersTrueStillAllowsProfile(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, v3MarkersPolicyPrefix+`

[agents.detection]
environment_markers = true

[profiles.`+gateHandoffProfile+`]
action = "allow"
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
	if !permitted {
		t.Fatalf("Decide = false, want true; stderr = %q", stderr.String())
	}
}

func TestDecideVersion3EnvironmentMarkersTrueStillDeniesProfile(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, v3MarkersPolicyPrefix+`

[agents.detection]
environment_markers = true

[profiles.`+gateTestProfile+`]
action = "deny"
`)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateProfileFlag,
			gateTestProfile,
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

	assertGateCancel(t, stderr.String())
}

func TestDecideVersion3ParentProcessCorruptPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, v3MarkersPolicyPrefix+`

[agents.detection]
environment_markers = true
parent_process = false
`)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())

	if !strings.Contains(stderr.String(), "the policy file is unusable") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
