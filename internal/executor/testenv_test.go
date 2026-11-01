package executor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	handoffPolicyBody = "version = 1\n\n[defaults]\naction = \"allow\"\n\n" +
		"[credential_export]\naction = \"allow\"\n"
	cancellationSecretArg = "SECRET_ARG_TOKEN_must_not_leak"
)

func policyConfigEnv(t *testing.T, work string) string {
	t.Helper()

	dir := filepath.Join(work, "guard-config")
	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte(handoffPolicyBody), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return "AWS_CLI_GUARD_CONFIG_DIR=" + dir + "\x00AWS_CLI_GUARD_STATE_DIR=" + stateDir
}

func policyConfigEnvEntries(t *testing.T, work string) []string {
	t.Helper()

	joined := policyConfigEnv(t, work)

	return strings.Split(joined, "\x00")
}

func guardDirEnvEntries(t *testing.T, work string) []string {
	t.Helper()

	configDir := filepath.Join(work, "guard-config")
	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	return []string{
		"AWS_CLI_GUARD_CONFIG_DIR=" + configDir,
		"AWS_CLI_GUARD_STATE_DIR=" + stateDir,
	}
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
