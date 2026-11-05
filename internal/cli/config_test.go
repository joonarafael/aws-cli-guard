package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

func TestRunConfigValidateSuccess(t *testing.T) {
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
		[]string{progName, configSubcommand, configValidateSub},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	want := "aws-cli-guard: policy: valid\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunConfigValidateSuccessV3(t *testing.T) {
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
environment_markers = false
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, configSubcommand, configValidateSub},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	want := "aws-cli-guard: policy: valid\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunConfigValidateDoesNotPrintActions(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "allow"

[credential_export]
action = "deny"

[profiles.secret]
action = "deny"
`, 0o600)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, configSubcommand, configValidateSub},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutNotContains(t, stdout.String(), "defaults:")
	assertStdoutNotContains(t, stdout.String(), "profile:")
}

func TestRunConfigValidateMissingPolicy(t *testing.T) {
	t.Parallel()

	env := policyTestEnv(t)

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, configSubcommand, configValidateSub},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
	)
	assertPolicyCancel(t, status, stderr.String())
}

func TestRunConfigUsage(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{progName, configSubcommand},
		{progName, configSubcommand, "unknown"},
		{progName, configSubcommand, configValidateSub, "extra"},
	}

	for _, argv := range cases {
		var stderr bytes.Buffer

		status := cli.Dispatch(argv, nil, &bytes.Buffer{}, &stderr, panicIfConfirm)
		if status != output.CancelStatus {
			t.Fatalf("argv %v status = %d", argv, status)
		}

		if !stringsContainsUsageConfig(stderr.String()) {
			t.Fatalf("argv %v stderr = %q", argv, stderr.String())
		}
	}
}

func stringsContainsUsageConfig(stderr string) bool {
	return strings.Contains(stderr, "usage: aws-guard config validate")
}
