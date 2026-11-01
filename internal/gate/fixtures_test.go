package gate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	gateSecretArg           = "SECRET_ARG_TOKEN_gate_test"
	gateTestProfile         = "production-admin"
	gateHandoffProfile      = "staging-handoff-test"
	gateProgName            = "aws"
	gateProfileFlag         = "--profile"
	gateSTSService          = "sts"
	gateGetCallerIdentityOp = "get-caller-identity"
	gateSSOCommand          = "sso"
	gateLoginOperation      = "login"
	gateLogoutOperation     = "logout"
)

func gateEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	return []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_CONFIG_DIR=" + filepath.Join(home, "config"),
		"AWS_CLI_GUARD_STATE_DIR=" + filepath.Join(home, "state"),
		"AWS_CONFIG_FILE=" + filepath.Join(home, "aws-config-empty"),
		"AWS_SHARED_CREDENTIALS_FILE=" + filepath.Join(home, "aws-creds-empty"),
	}
}

func writeGatePolicy(t *testing.T, env []string, body string) {
	t.Helper()

	dir := configDirFromEnv(t, env)

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

func writeGateGrants(t *testing.T, env []string, body string) {
	t.Helper()

	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "grants.json")

	err = os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func configDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	return dirFromEnv(t, env, "AWS_CLI_GUARD_CONFIG_DIR=")
}

func stateDirFromEnv(t *testing.T, env []string) string {
	t.Helper()

	return dirFromEnv(t, env, "AWS_CLI_GUARD_STATE_DIR=")
}

func dirFromEnv(t *testing.T, env []string, prefix string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, prefix); ok {
			return after
		}
	}

	t.Fatalf("%s missing from env", prefix)

	return ""
}

func policyPathFromEnv(t *testing.T, env []string) string {
	t.Helper()

	return filepath.Join(configDirFromEnv(t, env), "config.toml")
}

func allowPolicyBody(profile string) string {
	return `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "allow"

[profiles.` + profile + `]
action = "allow"
`
}

func denyPolicyBody(profile string) string {
	return `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "allow"

[profiles.` + profile + `]
action = "deny"
`
}

func promptPolicyBody(profile string) string {
	return `version = 1

[defaults]
action = "deny"

[credential_export]
action = "allow"

[profiles.` + profile + `]
action = "prompt"
`
}

func exportPromptPolicyBody(profile string) string {
	return `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"

[profiles.` + profile + `]
action = "allow"
`
}

func currentGrantBody(profile string) string {
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
}`
}
