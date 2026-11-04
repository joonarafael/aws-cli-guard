package cli_test

import (
	"aws-cli-guard/internal/cli"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunPolicyExplainProfileOverride(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"

[profiles.production-admin]
action = "allow"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, policyTestProfile},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: action: allow\n")
	assertStdoutContains(t, out, "aws-cli-guard: basis: profile\n")
	assertStdoutContains(t, out, "aws-cli-guard: grant: not used\n")
	assertStdoutContains(t, out, "aws-cli-guard: credential_export: prompt\n")
}

func TestRunPolicyExplainDefaultsForUnknownProfile(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "deny"

[credential_export]
action = "allow"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, "unknown-profile"},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: action: deny\n")
	assertStdoutContains(t, out, "aws-cli-guard: basis: defaults\n")
	assertStdoutContains(t, out, "aws-cli-guard: grant: not used\n")
}

func TestRunPolicyExplainAccountAndRole(t *testing.T) {
	t.Parallel()

	env, awsConfigPath := policyTestEnvWithAWSConfig(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"
`, 0o600)

	configBody := strings.Join([]string{
		"[profile " + policyTestProfile + "]",
		"account_id = " + policyExplainAccount,
		"role_arn = " + policyExplainRoleARN,
	}, "\n")

	err := os.WriteFile(awsConfigPath, []byte(configBody), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, policyTestProfile},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: account: "+policyExplainAccount)
	assertStdoutContains(t, out, "aws-cli-guard: role: "+policyExplainRoleARN)
	assertStdoutContains(t, out, "aws-cli-guard: grant: required\n")
	assertDetailOrder(t, out, "profile:", "account:", "role:", "action:")
}

func TestRunPolicyExplainCredentialProcessIdentityLine(t *testing.T) {
	t.Parallel()

	env, awsConfigPath := policyTestEnvWithAWSConfig(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"
`, 0o600)

	configBody := strings.Join([]string{
		"[profile " + policyTestProfile + "]",
		"credential_process = echo " + policyCredProcMarker,
	}, "\n")

	err := os.WriteFile(awsConfigPath, []byte(configBody), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, policyTestProfile},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "identity: configured source only; the caller was not verified")
	assertStdoutNotContains(t, out, policyCredProcMarker)
}

func TestRunPolicyExplainFingerprintFailureStillSuccess(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"
`, 0o600)

	// AWS_CONFIG_FILE points at a directory so Load fails.
	env = append(env, "AWS_CONFIG_FILE="+configDirFromPolicyEnv(t, env))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, policyTestProfile},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: action: prompt\n")
	assertStdoutNotContains(t, out, "account:")
}

func TestRunPolicyExplainApprovalTTL(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"

[profiles.production-admin]
action = "prompt"
approval_ttl = "10m"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, policyTestProfile},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: approval_ttl: 10m0s\n")
}
