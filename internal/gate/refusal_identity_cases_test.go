package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDecideDenySourceProfileChainShowsIdentityLine(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile base]",
		"account_id = 999999999999",
		"role_arn = " + identityRoleARN,
		"",
		"[profile " + gateTestProfile + "]",
		"source_profile = base",
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, identityUnverifiedDetail)
	assertNotContains(t, stderr, "aws-cli-guard: role:")
	assertNotContains(t, stderr, "999999999999")
	assertNotContains(t, stderr, "profile: base")
}

func TestDecideDenyCredentialProcessOmitsCommand(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"credential_process = " + identityCredProcMarker,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, identityUnverifiedDetail)
	assertNotContains(t, stderr, identityCredProcMarker)
}

func TestDecideDenyCredentialSourceNotEchoed(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"credential_source = Ec2InstanceMetadata",
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, identityUnverifiedDetail)
	assertNotContains(t, stderr, "Ec2InstanceMetadata")
}

func TestDecideDenyInvalidAccountControlOmitsIdentityLines(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = 123\t456789012",
		"role_arn = " + identityRoleARN,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertNotContains(t, stderr, "aws-cli-guard: account:")
	assertNotContains(t, stderr, "aws-cli-guard: role:")
	assertNotContains(t, stderr, "456789012")
}

func TestDecideDenyUnreadableAWSConfigOmitsIdentity(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))

	configPath := awsConfigPathFromEnv(t, env)

	err := os.Mkdir(configPath, 0o700)
	if err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, "policy: deny")
	assertNotContains(t, stderr, "aws-cli-guard: account:")
	assertNotContains(t, stderr, identityUnverifiedDetail)
}

func TestDecideDenyMissingProfileOmitsIdentity(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, "[profile other]\nregion = us-east-1\n")

	stderr := decideDenied(t, env, denyArgs())

	assertNotContains(t, stderr, "aws-cli-guard: account:")
	assertNotContains(t, stderr, "aws-cli-guard: role:")
	assertNotContains(t, stderr, identityUnverifiedDetail)
}

func TestDecidePromptWithoutGrantShowsAccountAndAllowGuidance(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = " + identityAccountID,
		"role_arn = " + identityRoleARN,
	}, "\n"))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "ec2", "describe-instances", gateProfileFlag, gateTestProfile},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	out := stderr.String()
	assertContains(t, out, "aws-cli-guard: account: "+identityAccountID)
	assertContains(t, out, "aws-guard allow "+gateTestProfile)
}

func TestDecideExportRefusalOmitsIdentity(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, exportPromptPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = " + identityAccountID,
	}, "\n"))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateConfigureCommand,
			gateExportCredentialsOp,
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
	assertNotContains(t, out, "aws-cli-guard: account:")
	assertNotContains(t, out, identityUnverifiedDetail)
}

func TestDecideEnvironmentRefusalOmitsIdentity(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "allow"
`)
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = " + identityAccountID,
	}, "\n"))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		append(env,
			"AWS_ACCESS_KEY_ID=AKIAEXAMPLE",
			"AWS_SECRET_ACCESS_KEY=secret",
		),
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	out := stderr.String()
	assertNotContains(t, out, "aws-cli-guard: account:")
	assertNotContains(t, out, identityUnverifiedDetail)
}
