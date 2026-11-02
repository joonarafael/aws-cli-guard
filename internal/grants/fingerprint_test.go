package grants_test

import (
	"aws-cli-guard/internal/grants"
	"errors"
	"strings"
	"testing"
)

func TestFingerprintAssumeRoleProfile(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile admin]",
		"role_arn = " + fixtureRoleARN,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "admin")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.RoleARN != fixtureRoleARN {
		t.Fatalf("RoleARN = %q", got.RoleARN)
	}

	if got.AccountID != "" || got.SSOAccountID != "" || got.SSORoleName != "" ||
		got.CredentialProcess {
		t.Fatalf("source: %+v", got)
	}
}

func TestFingerprintExplicitAccountID(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile acct]",
		"account_id = 123456789012",
		"role_arn = " + fixtureRoleARN,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "acct")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.AccountID != "123456789012" {
		t.Fatalf("AccountID = %q", got.AccountID)
	}
}

func TestFingerprintSSOProfileFields(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[sso-session corp]",
		"sso_account_id = 999999999999",
		"sso_role_name = SessionOnly",
		"",
		"[profile app]",
		"sso_session = corp",
		"sso_account_id = 111111111111",
		"sso_role_name = Developer",
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "app")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.SSOAccountID != "111111111111" || got.SSORoleName != "Developer" {
		t.Fatalf("source: %+v", got)
	}
}

func TestFingerprintCredentialProcessFlag(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile proc]",
		"credential_process = " + fixtureCredProcFP,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "proc")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if !got.CredentialProcess {
		t.Fatal("expected credential_process flag")
	}

	if strings.Contains(got.RoleARN+got.AccountID, fixtureCredProcFP) {
		t.Fatal("command leaked into source fields")
	}
}

func TestFingerprintCredentialProcessInNestedBlock(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile nestedproc]",
		"s3 =",
		"    credential_process = " + fixtureCredProcFP,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "nestedproc")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if !got.CredentialProcess {
		t.Fatal("expected credential_process flag from nested block")
	}
}

func TestFingerprintBindsSourceProfileChain(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		fixtureProfileBaseHeader,
		"role_arn = " + fixtureRoleARN,
		"",
		"[profile chained]",
		fixtureSourceProfileBase,
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "chained")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.SourceProfile != fixtureProfileBaseName || len(got.Chain) != 1 {
		t.Fatalf("source: %+v", got)
	}

	if got.Chain[0].Profile != fixtureProfileBaseName || got.Chain[0].RoleARN != fixtureRoleARN {
		t.Fatalf("chain[0]: %+v", got.Chain[0])
	}
}

func TestFingerprintAmbiguousSourceProfileAndCredentialSource(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile bad]",
		"source_profile = base",
		"credential_source = Ec2InstanceMetadata",
		"",
		"[profile base]",
		"role_arn = " + fixtureRoleARN,
	}, "\n"), "")

	_, err := grants.Fingerprint(env, "bad")
	if err == nil {
		t.Fatal("expected resolve error")
	}

	grErr, ok := errors.AsType[*grants.Error](err)
	if !ok || grErr.Kind != grants.KindResolve {
		t.Fatalf("got %v", err)
	}
}

func TestFingerprintCredentialsOverrideConfig(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env,
		"[profile merge]\nrole_arn = arn:aws:iam::1:role/Config\n",
		"[merge]\nrole_arn = arn:aws:iam::2:role/Credentials\n",
	)

	got, err := grants.Fingerprint(env, "merge")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.RoleARN != "arn:aws:iam::2:role/Credentials" {
		t.Fatalf("RoleARN = %q", got.RoleARN)
	}
}

func TestFingerprintDefaultInheritance(t *testing.T) {
	t.Parallel()

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[DEFAULT]",
		"role_arn = " + fixtureRoleARN,
		"",
		"[profile child]",
		"region = us-east-1",
	}, "\n"), "")

	got, err := grants.Fingerprint(env, "child")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	if got.RoleARN != fixtureRoleARN {
		t.Fatalf("RoleARN = %q", got.RoleARN)
	}
}
