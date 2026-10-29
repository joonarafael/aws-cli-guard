package grants_test

import (
	"aws-cli-guard/internal/awsconfig"
	"aws-cli-guard/internal/grants"
	"errors"
	"strings"
	"testing"
)

func TestFingerprintDefaultProfileName(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[default]",
		"role_arn = " + fixtureRoleARN,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "default")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.RoleARN != fixtureRoleARN {
		t.Fatalf("RoleARN = %q", got.RoleARN)
	}
}

func TestFingerprintMissingProfileEmpty(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, "[profile other]\nregion = us-east-1\n", "")

	got, err := grants.Fingerprint(env, "absent")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got != emptySource() {
		t.Fatalf("source: %+v", got)
	}
}

func TestFingerprintMissingAWSFilesEmpty(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)

	got, err := grants.Fingerprint(env, "any")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got != emptySource() {
		t.Fatalf("source: %+v", got)
	}
}

func TestFingerprintUnreadableConfigPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{
		"HOME=" + dir,
		"AWS_CONFIG_FILE=" + dir,
	}

	_, err := grants.Fingerprint(env, "p")
	if err == nil {
		t.Fatal("expected read error")
	}

	if !errors.Is(err, awsconfig.ErrNotRegular) {
		t.Fatalf("errors.Is ErrNotRegular = false, got %v", err)
	}
}

func TestFingerprintEmptyProfile(t *testing.T) {
	t.Parallel()

	_, err := grants.Fingerprint(nil, "")
	if !errors.Is(err, grants.ErrEmptyProfile) {
		t.Fatalf("got %v", err)
	}
}

func TestFingerprintContinuedRoleARNInvalid(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile bad]",
		"role_arn = arn:aws:iam::123:role/Start",
		"    continued-secret-suffix",
	}, "\n"), "")

	_, err := grants.Fingerprint(env, "bad")
	if err == nil {
		t.Fatal("expected invalid value")
	}

	grErr, ok := errors.AsType[*grants.Error](err)
	if !ok || grErr.Kind != grants.KindInvalid || grErr.Key != "source.role_arn" {
		t.Fatalf("got %v", err)
	}

	if strings.Contains(err.Error(), "continued-secret-suffix") {
		t.Fatalf("error leaks value: %v", err)
	}
}

func TestSourceEqual(t *testing.T) {
	t.Parallel()

	left := grants.Source{
		AccountID:         "1",
		RoleARN:           fixtureRoleARN,
		SSOAccountID:      "2",
		SSORoleName:       "Dev",
		CredentialProcess: true,
	}

	right := left
	if !left.Equal(right) {
		t.Fatal("Equal copy")
	}

	if !emptySource().Equal(emptySource()) {
		t.Fatal("empty sources should match")
	}

	right.RoleARN = "other"
	if left.Equal(right) {
		t.Fatal("expected role_arn mismatch")
	}

	right = left

	right.CredentialProcess = false
	if left.Equal(right) {
		t.Fatal("expected credential_process mismatch")
	}
}
