package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	revokeSubcommand = "revoke"
	revokeAllFlag    = "--all"
)

func revokeTestEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	return []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_STATE_DIR=" + filepath.Join(home, "state"),
	}
}

func revokeStateDir(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "AWS_CLI_GUARD_STATE_DIR="); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_STATE_DIR missing")

	return ""
}

func writeRevokeGrants(t *testing.T, env []string, body string) {
	t.Helper()

	dir := revokeStateDir(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "grants.json"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func panicIfConfirm(string, string) error {
	panic("revoke must not confirm")
}

func sampleRevokeGrantJSON(profile string) string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "` + profile + `",
      "expires": "2099-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    }
  ]
}
`
}
