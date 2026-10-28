package grants_test

import (
	"aws-cli-guard/internal/grants"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureSecret = "AKIAIOSFODNN7EXAMPLE-grants-body"

func emptySource() grants.Source {
	return grants.Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
	}
}

func stateDirEnv(t *testing.T) []string {
	t.Helper()

	dir := t.TempDir()

	return []string{"AWS_CLI_GUARD_STATE_DIR=" + dir}
}

func writeGrants(t *testing.T, env []string, body string) {
	t.Helper()

	writeGrantsPerm(t, env, body, 0o600)
}

func writeGrantsPerm(t *testing.T, env []string, body string, perm os.FileMode) {
	t.Helper()

	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "grants.json")

	err = os.WriteFile(path, []byte(body), perm)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func grantsPathFromEnv(t *testing.T, env []string) string {
	t.Helper()

	dir := stateDirFromEnv(t, env)

	return filepath.Join(dir, "grants.json")
}

func stateDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_STATE_DIR="
		if len(entry) > len(prefix) && entry[:len(prefix)] == prefix {
			return entry[len(prefix):]
		}
	}

	t.Fatal("AWS_CLI_GUARD_STATE_DIR not in env")

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

func sampleGrantJSON(profile, expires string) string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "` + profile + `",
      "expires": "` + expires + `",
      "source": {
        "account_id": "123456789012",
        "role_arn": "arn:aws:iam::123456789012:role/Admin",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    }
  ]
}
`
}
