//go:build awsguard_testpaths

package gate_test

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/gate"
	"bytes"
	"strings"
	"testing"
)

const v4AgentsPolicyPrefix = `version = 4

[defaults]
action = "allow"

[environment]
action = "allow"

[credential_export]
action = "allow"

[agents.detection]
environment_markers = true
`

func TestDecideAgentDenyOverridesProfileAllow(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	env = append(env, agent.TestAgentMarkerEnv+"=1")
	writeGatePolicy(t, env, v4AgentsPolicyPrefix+`

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

	if !strings.Contains(out, gateHandoffProfile+" is denied for recognized agents") {
		t.Fatalf("stderr = %q", out)
	}
}

func TestDecideAgentAllowOverridesProfilePrompt(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	env = append(env, agent.TestAgentMarkerEnv+"=1")
	writeGatePolicy(t, env, v4AgentsPolicyPrefix+`

[profiles.`+gateHandoffProfile+`]
action = "prompt"

[agents.profiles.`+gateHandoffProfile+`]
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
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, want true; stderr = %q", stderr.String())
	}
}

func TestDecideAgentRulesIgnoredWithoutMarker(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, v4AgentsPolicyPrefix+`

[profiles.`+gateTestProfile+`]
action = "allow"

[agents.profiles.`+gateTestProfile+`]
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
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, want true; stderr = %q", stderr.String())
	}
}

func TestDecideAgentDenyBeforeCredentialExport(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	env = append(env, agent.TestAgentMarkerEnv+"=1")
	writeGatePolicy(t, env, v4AgentsPolicyPrefix+`

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
			gateConfigureCommand,
			gateExportCredentialsOp,
		},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	out := stderr.String()
	if !strings.Contains(out, "is denied for recognized agents") {
		t.Fatalf("stderr = %q", out)
	}
}
