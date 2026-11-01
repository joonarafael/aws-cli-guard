package executor_test

import (
	"aws-cli-guard/internal/executor"
	"os"
	"path/filepath"
	"testing"
)

func enforceEnv(t *testing.T, work, fake, record string) []string {
	t.Helper()

	home := filepath.Join(work, "home")
	awsConfig := filepath.Join(home, "empty-config")
	awsCreds := filepath.Join(home, "empty-creds")

	err := os.MkdirAll(home, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	return envHandoff(append(
		guardDirEnvEntries(t, work),
		"HOME="+home,
		"AWS_CONFIG_FILE="+awsConfig,
		"AWS_SHARED_CREDENTIALS_FILE="+awsCreds,
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
	)...)
}

func writeEnforcePolicy(t *testing.T, work, body string) {
	t.Helper()

	configDir := filepath.Join(work, "guard-config")

	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func writeEnforceGrants(t *testing.T, work, body string) {
	t.Helper()

	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(stateDir, "grants.json"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func denyProfilePolicy() string {
	return `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "allow"

[profiles.` + enforceProfile + `]
action = "deny"
`
}

func promptProfilePolicy() string {
	return `version = 1

[defaults]
action = "deny"

[credential_export]
action = "allow"

[profiles.` + enforceProfile + `]
action = "prompt"
`
}

func exportPromptAllowProfilePolicy() string {
	return `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"

[profiles.` + enforceProfile + `]
action = "allow"
`
}

func currentGrantJSON() string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "` + enforceProfile + `",
      "expires": "` + enforceGrantExpiry + `",
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
