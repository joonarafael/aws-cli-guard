package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

func TestRunAllowUnresolvedCredentialSourceBeforeConfirm(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowAWSConfig(t, env, strings.Join([]string{
		"[profile a]",
		"source_profile = b",
		"",
		"[profile b]",
		"source_profile = a",
	}, "\n"))

	var stderr bytes.Buffer

	called := false

	status := cli.Dispatch(
		allowArgvWithTTL("a"),
		env,
		&bytes.Buffer{},
		&stderr,
		func(string, string) error {
			called = true

			return nil
		},
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if called {
		t.Fatal("confirm should not run")
	}

	if !strings.Contains(stderr.String(), "the credential source could not be resolved") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	if strings.Contains(stderr.String(), "source_profile") {
		t.Fatalf("stderr leaks config: %q", stderr.String())
	}
}
