package awsconfig_test

import (
	"aws-cli-guard/internal/awsconfig"
	"errors"
	"strconv"
	"strings"
	"testing"
)

const (
	resolveRoleARN     = "arn:aws:iam::123456789012:role/Admin"
	resolveProfileLeaf = "[profile leaf]"
	resolveProfileApp  = "[profile app]"
)

func TestResolveProfileTwoHopChainOrder(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[profile root]",
		"role_arn = " + resolveRoleARN,
		"",
		"[profile mid]",
		"source_profile = root",
		"",
		resolveProfileLeaf,
		"source_profile = mid",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "leaf")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if got.SourceProfile != "mid" || len(got.Chain) != 2 {
		t.Fatalf("source: %+v", got)
	}

	if got.Chain[0].Profile != "mid" || got.Chain[1].Profile != "root" {
		t.Fatalf("chain order: %+v", got.Chain)
	}

	if got.Chain[1].RoleARN != resolveRoleARN {
		t.Fatalf("root role: %+v", got.Chain[1])
	}
}

func TestResolveProfileSSOSessionWinsOverProfileURL(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[sso-session corp]",
		"sso_start_url = https://session.example/start",
		"sso_region = us-west-2",
		"",
		resolveProfileApp,
		"sso_session = corp",
		"sso_start_url = https://profile.example/start",
		"sso_region = us-east-1",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "app")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if got.SSOStartURL != "https://session.example/start" || got.SSORegion != "us-west-2" {
		t.Fatalf("source: %+v", got)
	}
}

func TestResolveProfileInlineSSOURLWhenNoSession(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		resolveProfileApp,
		"sso_start_url = https://profile.example/start",
		"sso_region = eu-central-1",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "app")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if got.SSOStartURL != "https://profile.example/start" || got.SSORegion != "eu-central-1" {
		t.Fatalf("source: %+v", got)
	}
}

func TestResolveProfileSessionAccountIDDoesNotReplaceProfileField(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[sso-session corp]",
		"sso_account_id = 999999999999",
		"",
		resolveProfileApp,
		"sso_session = corp",
		"sso_account_id = 111111111111",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "app")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if got.SSOAccountID != "111111111111" {
		t.Fatalf("SSOAccountID = %q", got.SSOAccountID)
	}
}

func TestResolveProfileCredentialSourceTerminal(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[profile base]",
		"role_arn = " + resolveRoleARN,
		"",
		resolveProfileLeaf,
		"credential_source = Ec2InstanceMetadata",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "leaf")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if got.CredentialSource != "Ec2InstanceMetadata" || len(got.Chain) != 0 {
		t.Fatalf("source: %+v", got)
	}
}

func TestResolveProfileMissingParent(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, resolveProfileLeaf+"\nsource_profile = missing\n")

	_, err := awsconfig.ResolveProfile(cfg, "leaf")
	assertResolveError(t, err, "source_profile")
}

func TestResolveProfileMissingSSOSession(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, resolveProfileApp+"\nsso_session = missing\n")

	_, err := awsconfig.ResolveProfile(cfg, "app")
	assertResolveError(t, err, "sso_session")
}

func TestResolveProfileCycle(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[profile a]",
		"source_profile = b",
		"",
		"[profile b]",
		"source_profile = a",
	}, "\n"))

	_, err := awsconfig.ResolveProfile(cfg, "a")
	assertResolveError(t, err, "cycle")
}

func TestResolveProfileMaxDepth(t *testing.T) {
	t.Parallel()

	lines := make([]string, 0, 40)
	for i := range 9 {
		lines = append(lines,
			"[profile p"+strconv.Itoa(i)+"]",
			"source_profile = p"+strconv.Itoa(i+1),
			"",
		)
	}

	lines = append(lines, "[profile p9]", "role_arn = "+resolveRoleARN)

	cfg := loadResolveConfig(t, strings.Join(lines, "\n"))

	_, err := awsconfig.ResolveProfile(cfg, "p0")
	assertResolveError(t, err, "depth")
}

func TestResolveProfileEightHopsOK(t *testing.T) {
	t.Parallel()

	lines := []string{
		"[profile p0]",
		"source_profile = p1",
		"",
		"[profile p1]",
		"source_profile = p2",
		"",
		"[profile p2]",
		"source_profile = p3",
		"",
		"[profile p3]",
		"source_profile = p4",
		"",
		"[profile p4]",
		"source_profile = p5",
		"",
		"[profile p5]",
		"source_profile = p6",
		"",
		"[profile p6]",
		"source_profile = p7",
		"",
		"[profile p7]",
		"source_profile = p8",
		"",
		"[profile p8]",
		"role_arn = " + resolveRoleARN,
	}

	cfg := loadResolveConfig(t, strings.Join(lines, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "p0")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if len(got.Chain) != 8 {
		t.Fatalf("chain len = %d", len(got.Chain))
	}
}

func TestResolveProfileExternalIDOnHop(t *testing.T) {
	t.Parallel()

	cfg := loadResolveConfig(t, strings.Join([]string{
		"[profile base]",
		"role_arn = " + resolveRoleARN,
		"external_id = hop-secret-token",
		"",
		resolveProfileLeaf,
		"source_profile = base",
	}, "\n"))

	got, err := awsconfig.ResolveProfile(cfg, "leaf")
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}

	if len(got.Chain) != 1 || got.Chain[0].ExternalID != "hop-secret-token" {
		t.Fatalf("chain: %+v", got.Chain)
	}

	if strings.Contains(got.Chain[0].ExternalID+got.Chain[0].RoleARN, "UNEXPECTED") {
		t.Fatal("unexpected leak marker")
	}
}

func loadResolveConfig(t *testing.T, configBody string) awsconfig.Config {
	t.Helper()

	home := t.TempDir()
	env := []string{"HOME=" + home}
	writeConfig(t, home, configBody, "")

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return cfg
}

func assertResolveError(t *testing.T, err error, wantKey string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}

	cfgErr, ok := errors.AsType[*awsconfig.Error](err)
	if !ok || cfgErr.Kind != awsconfig.KindResolve || cfgErr.Key != wantKey {
		t.Fatalf("got %v", err)
	}

	if strings.Contains(err.Error(), "missing") || strings.Contains(err.Error(), "https://") {
		t.Fatalf("error leaks config detail: %v", err)
	}
}
