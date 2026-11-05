package cli_test

import (
	"aws-cli-guard/internal/cli"
	"bytes"
	"testing"
)

func TestRunPolicyListV4AgentRules(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 4

[defaults]
action = "prompt"

[environment]
action = "prompt"

[credential_export]
action = "prompt"

[agents.detection]
environment_markers = true

[agents]
default = "deny"

[agents.profiles.dev-readonly]
action = "allow"
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
	assertStdoutContains(t, out, "aws-cli-guard: agents_default: deny\n")
	assertStdoutContains(t, out, "aws-cli-guard: agent_profile: dev-readonly\n")
	assertStdoutContains(t, out, "aws-cli-guard: agent_action: allow\n")
}

func TestRunPolicyExplainV4AgentOverride(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 4

[defaults]
action = "prompt"

[environment]
action = "prompt"

[credential_export]
action = "prompt"

[agents.detection]
environment_markers = true

[profiles.production-admin]
action = "allow"

[agents.profiles.production-admin]
action = "deny"
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
	assertStdoutContains(t, out, "aws-cli-guard: agent_action: deny\n")
	assertStdoutContains(t, out, "aws-cli-guard: agent_basis: agents.profile\n")
}
