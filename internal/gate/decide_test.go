package gate_test

import (
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/policy"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDecideMissingPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateSecretArg},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), policy.ReasonMissing)
	assertNoSecret(t, stderr.String())
}

func TestDecideAllowProfile(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
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
		append(env, "AWS_PROFILE=ignored"),
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}
}

func TestDecideDenyBeatsGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
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
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "policy: deny")
	assertNoSecret(t, stderr.String())
}

func TestDecidePromptWithoutGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "ec2", "describe-instances", gateProfileFlag, gateTestProfile},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "aws-guard allow "+gateTestProfile)
	assertContains(t, stderr.String(), "command: ec2 describe-instances")
}

func TestDecidePromptWithCurrentGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))
	writeGateGrants(t, env, currentGrantBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "s3", "ls", gateProfileFlag, gateTestProfile},
		env,
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}
}

func TestDecideExportRefused(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, exportPromptPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateConfigureCommand,
			"export-credentials",
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), policy.ReasonExportRefused)

	if strings.Contains(stderr.String(), "configure") {
		t.Fatal("export refusal must not name the command")
	}
}

func TestDecideEnvironmentNotCoveredByGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "allow"
`)
	writeGateGrants(t, env, currentGrantBody(gateTestProfile))

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

	assertContains(t, stderr.String(), "a grant cannot authorize this subject")
}

func TestDecideCorruptPolicyUnchanged(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, "not valid toml\n")

	path := policyPathFromEnv(t, env)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "the policy file is unusable")
	assertContains(t, stderr.String(), "aws-guard recover policy")

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(after) != string(before) {
		t.Fatal("corrupt policy file was modified")
	}
}

func TestDecideDefaultProfileDeniedIncludesWarning(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, `version = 1

[defaults]
action = "deny"

[credential_export]
action = "allow"
`)

	var stderr bytes.Buffer

	permitted := gate.Decide([]string{gateProgName, "s3", "ls"}, env, &stderr)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "aws-cli-guard: warning:")
	assertContains(t, stderr.String(), "default profile")
	assertContains(t, stderr.String(), "aws-cli-guard: error:")
}
