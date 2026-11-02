package gate_test

import (
	"strings"
	"testing"
)

func TestDecideDenyShowsAccountAndRole(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = " + identityAccountID,
		"role_arn = " + identityRoleARN,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, "aws-cli-guard: account: "+identityAccountID)
	assertContains(t, stderr, "aws-cli-guard: role: "+identityRoleARN)
	assertDetailOrder(t, stderr, "profile:", "account:", "role:", "command:")
	assertNotContains(t, stderr, identityUnverifiedDetail)
}

func TestDecideDenyRoleARNOnlyNoAccountLine(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"role_arn = " + identityRoleARN,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, "aws-cli-guard: role: "+identityRoleARN)
	assertNotContains(t, stderr, "aws-cli-guard: account:")
}

func TestDecideDenySSOProfileWithoutIdentityLine(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[sso-session corp]",
		"sso_start_url = " + identitySessionStartURL,
		"sso_region = us-east-1",
		"",
		"[profile " + gateTestProfile + "]",
		"sso_session = corp",
		"sso_account_id = " + identitySSOAccountID,
		"sso_role_name = " + identitySSORoleName,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, "aws-cli-guard: account: "+identitySSOAccountID)
	assertContains(t, stderr, "aws-cli-guard: role: "+identitySSORoleName)
	assertNotContains(t, stderr, identitySessionStartURL)
	assertNotContains(t, stderr, identityUnverifiedDetail)
}

func TestDecideDenyStaticAndSSOKeys(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateAWSConfig(t, env, strings.Join([]string{
		"[profile " + gateTestProfile + "]",
		"account_id = " + identityAccountID,
		"sso_account_id = " + identitySSOAccountID,
		"role_arn = " + identityRoleARN,
		"sso_role_name = " + identitySSORoleName,
	}, "\n"))

	stderr := decideDenied(t, env, denyArgs())

	assertContains(t, stderr, "aws-cli-guard: account: "+identityAccountID)
	assertContains(t, stderr, "aws-cli-guard: sso account: "+identitySSOAccountID)
	assertContains(t, stderr, "aws-cli-guard: role: "+identityRoleARN)
	assertContains(t, stderr, "aws-cli-guard: sso role: "+identitySSORoleName)
	assertDetailOrder(
		t,
		stderr,
		"profile:",
		"account:",
		"sso account:",
		"role:",
		"sso role:",
		"command:",
	)
}
