package policy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSecret = "AKIAIOSFODNN7EXAMPLE-policy-body"

func configDirEnv(t *testing.T) []string {
	t.Helper()

	dir := t.TempDir()

	return []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}
}

func writePolicy(t *testing.T, env []string, body string) {
	t.Helper()

	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.toml")

	err = os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func writePolicyPerm(t *testing.T, env []string, body string, perm os.FileMode) {
	t.Helper()

	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.toml")

	err = os.WriteFile(path, []byte(body), perm)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func policyDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_CONFIG_DIR="
		if len(entry) > len(prefix) && entry[:len(prefix)] == prefix {
			return entry[len(prefix):]
		}
	}

	t.Fatal("AWS_CLI_GUARD_CONFIG_DIR not in env")

	return ""
}

func assertNoSecretInError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}

	if strings.Contains(err.Error(), fixtureSecret) {
		t.Fatalf("error must not contain fixture secret: %v", err)
	}
}
