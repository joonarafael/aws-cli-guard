package executor_test

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	minimalPolicyBody     = "version = 1\n"
	cancellationSecretArg = "SECRET_ARG_TOKEN_must_not_leak"
)

func policyConfigEnv(t *testing.T, work string) string {
	t.Helper()

	dir := filepath.Join(work, "guard-config")

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte(minimalPolicyBody), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return "AWS_CLI_GUARD_CONFIG_DIR=" + dir
}

func emptyPolicyConfigEnv(t *testing.T, work string) string {
	t.Helper()

	dir := filepath.Join(work, "empty-guard-config")

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	return "AWS_CLI_GUARD_CONFIG_DIR=" + dir
}
