package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/policy"
	"bytes"
	"strings"
	"testing"
)

func TestRunPolicyListSuccessV2(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 2

[defaults]
action = "prompt"

[environment]
action = "prompt"

[credential_export]
action = "prompt"

[profiles.dev-readonly]
action = "allow"

[profiles.production-admin]
action = "prompt"
approval_ttl = "5m"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: version: 2\n")
	assertStdoutContains(t, out, "aws-cli-guard: defaults: prompt\n")
	assertStdoutContains(t, out, "aws-cli-guard: environment: prompt\n")
	assertStdoutContains(t, out, "aws-cli-guard: environment_basis: environment\n")
	assertDetailOrder(t, out, "profile: dev-readonly", "profile: production-admin")
	assertStdoutContains(t, out, "aws-cli-guard: approval_ttl: 5m0s\n")
}

func TestRunPolicyListV3EnvironmentMarkersLine(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 3

[defaults]
action = "prompt"

[environment]
action = "prompt"

[credential_export]
action = "prompt"

[agents.detection]
environment_markers = true
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: version: 3\n")
	assertStdoutContains(t, stdout.String(), "aws-cli-guard: environment_markers: true\n")
}

func TestRunPolicyListV2OmitsEnvironmentMarkersLine(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 2

[defaults]
action = "prompt"

[environment]
action = "prompt"

[credential_export]
action = "prompt"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutNotContains(t, stdout.String(), "environment_markers:")
}

func TestRunPolicyListV1EnvironmentBasisDefaults(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "allow"

[credential_export]
action = "deny"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: environment_basis: defaults\n")
	assertStdoutNotContains(t, out, "profile:")
}

func TestRunPolicyListMissingPolicy(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
	)
	assertPolicyCancel(t, status, stderr.String())
	assertStderrContains(t, stderr.String(), policy.ReasonMissing)
}

func TestRunPolicyListCorruptPolicy(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, "not valid toml {{{", 0o600)

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
	)
	assertPolicyCancel(t, status, stderr.String())
	assertStderrContains(t, stderr.String(), "the policy file is unusable")
	assertStderrContains(t, stderr.String(), "aws-guard recover policy")
}

func TestRunPolicyListLooseModeWarns(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, "version = 1\n\n[defaults]\naction = \"prompt\"\n", 0o644)

	var stdout, stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyListSubcommand},
		env,
		&stdout,
		&stderr,
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStderrContains(t, stderr.String(), "aws-cli-guard: warning:")
	assertStdoutContains(t, stdout.String(), "aws-cli-guard: version: 1\n")
}

func TestRunPolicyUsage(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{progName, policySubcommand},
		{progName, policySubcommand, "unknown"},
		{progName, policySubcommand, policyListSubcommand, "extra"},
		{progName, policySubcommand, policyExplainSub},
		{progName, policySubcommand, policyExplainSub, "--bad"},
	}

	for _, argv := range cases {
		var stderr bytes.Buffer

		status := cli.Dispatch(argv, nil, &bytes.Buffer{}, &stderr, panicIfConfirm)
		assertPolicyCancel(t, status, stderr.String())

		if !stringsContainsUsagePolicy(stderr.String()) {
			t.Fatalf("argv %v stderr = %q", argv, stderr.String())
		}
	}
}

func TestRunPolicyExplainInvalidProfileName(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, "version = 1\n\n[defaults]\naction = \"prompt\"\n", 0o600)

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, policySubcommand, policyExplainSub, "bad\x01name"},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
	)
	assertPolicyCancel(t, status, stderr.String())
	assertStderrContains(t, stderr.String(), "the profile name is invalid")
}

func stringsContainsUsagePolicy(stderr string) bool {
	return strings.Contains(stderr, "usage: aws-guard policy list|explain <profile>")
}
