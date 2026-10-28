package awsconfig_test

import (
	"aws-cli-guard/internal/awsconfig"
	"strings"
	"testing"
)

func TestSecretsRedactedFromProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	body := strings.Join([]string{
		"[profile secrets]",
		"aws_access_key_id = " + fixtureAccessKey,
		"aws_secret_access_key = " + fixtureSecretKey,
		"aws_session_token = " + fixtureSession,
		"credential_process = " + fixtureCredProc,
		"region = eu-central-1",
	}, "\n")

	writeConfig(t, dir, body, "")

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	sec := cfg.Profiles["secrets"]
	assertSectionOmitsSecrets(t, sec)

	if !sec.HasAccessKeyID || !sec.HasSecretAccessKey || !sec.HasSessionToken {
		t.Fatal("expected access key flags")
	}

	if !sec.HasCredentialProcess {
		t.Fatal("expected credential_process flag")
	}

	if sec.Settings["region"] != "eu-central-1" {
		t.Fatalf("region = %q", sec.Settings["region"])
	}
}

func assertSectionOmitsSecrets(t *testing.T, sec awsconfig.Section) {
	t.Helper()

	blob := settingsBlob(sec)
	secrets := []string{fixtureAccessKey, fixtureSecretKey, fixtureSession, fixtureCredProc}

	for _, secret := range secrets {
		if strings.Contains(blob, secret) {
			t.Fatalf("section contains %q", secret)
		}
	}
}

func settingsBlob(sec awsconfig.Section) string {
	var parts []string

	for key, val := range sec.Settings {
		parts = append(parts, key+val)
	}

	for block, nested := range sec.Nested {
		parts = append(parts, block)

		for key, val := range nested {
			parts = append(parts, key+val)
		}
	}

	return strings.Join(parts, "")
}
