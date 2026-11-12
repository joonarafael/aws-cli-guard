package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"strings"
	"testing"
)

func TestDecideVerboseAllow(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=verbose")
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
			gateSecretArg,
		},
		env,
		"/usr/bin/fake-aws",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	out := stderr.String()
	assertNoSecret(t, out)

	for _, want := range []string{
		"subject: named profile " + gateTestProfile,
		"rule: profile allow",
		"stdin:",
		"caller: not recognized",
		"aws: /usr/bin/fake-aws",
		"outcome: allow",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stderr = %q, want substring %q", out, want)
		}
	}
}

func TestDecideVerboseDeny(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=verbose")
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	out := stderr.String()
	if !strings.Contains(out, "rule: profile deny") || !strings.Contains(out, "outcome: refuse") {
		t.Fatalf("stderr = %q", out)
	}
}

func TestDecideVerbosePromptNoGrant(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=verbose")
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	out := stderr.String()
	if !strings.Contains(out, "grant: absent") || !strings.Contains(out, "outcome: refuse") {
		t.Fatalf("stderr = %q", out)
	}
}

func TestDecideVerbosePromptWithGrant(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=verbose")
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))
	writeGateGrants(t, env, currentGrantBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	out := stderr.String()
	if !strings.Contains(out, "grant: current") || !strings.Contains(out, "outcome: allow") {
		t.Fatalf("stderr = %q", out)
	}
}

func TestDecideSettingsFileVerbose(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))
	writeGateSettingsVerbose(t, env)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if !strings.Contains(stderr.String(), "outcome: allow") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDecideEnvTightBeatsFileVerbose(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=tight")
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))
	writeGateSettingsVerbose(t, env)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatal("Decide = false")
	}

	if strings.Contains(stderr.String(), "outcome:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDecideInvalidLogEnv(t *testing.T) {
	t.Parallel()

	secret := "nonsense-level"
	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG="+secret)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	if strings.Contains(stderr.String(), secret) {
		t.Fatal("stderr echoed invalid env value")
	}

	assertGateCancel(t, stderr.String())
}

func TestDecideLoginInvalidLogEnv(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=bogus")

	var stderr bytes.Buffer

	permitted := gate.Decide(gateSSOLoginArgs(), env, "", &stderr)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
}
