package awsconfig_test

import (
	"aws-cli-guard/internal/awsconfig"
	"strings"
	"testing"
)

func TestLoadProfileSections(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	homeEnv := []string{"HOME=" + dir}

	configBody := strings.Join([]string{
		"[default]",
		"region = us-east-1",
		"",
		"[profile personal]",
		"region = eu-west-1",
		"",
		"[sso-session mysso]",
		"sso_region = us-east-1",
		"",
		"[preview]",
		"other = 1",
	}, "\n")

	writeConfig(t, dir, configBody, "")

	cfg, err := awsconfig.Load(homeEnv)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	def, found := cfg.Profiles["default"]
	if !found || def.Settings["region"] != "us-east-1" {
		t.Fatalf("default profile: %+v", def)
	}

	personal, found := cfg.Profiles["personal"]
	if !found || personal.Settings["region"] != "eu-west-1" {
		t.Fatalf("personal profile: %+v", personal)
	}

	if _, found := cfg.Profiles["preview"]; found {
		t.Fatal("preview must not be a profile")
	}

	sso, found := cfg.SSOSessions["mysso"]
	if !found || sso.Settings["sso_region"] != "us-east-1" {
		t.Fatalf("sso session: %+v", sso)
	}
}

func TestLoadCredentialsMerge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	configBody := strings.Join([]string{
		"[profile merged]",
		"region = from-config",
		"role_arn = arn:aws:iam::123:role/ReadOnly",
	}, "\n")

	credBody := strings.Join([]string{
		"[merged]",
		"aws_access_key_id = " + fixtureAccessKey,
		"aws_secret_access_key = " + fixtureSecretKey,
	}, "\n")

	writeConfig(t, dir, configBody, credBody)

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	sec := cfg.Profiles["merged"]
	if sec.Settings["region"] != "from-config" {
		t.Fatalf("region = %q", sec.Settings["region"])
	}

	if sec.Settings["role_arn"] != "arn:aws:iam::123:role/ReadOnly" {
		t.Fatalf("role_arn missing")
	}

	if !sec.HasAccessKeyID || !sec.HasSecretAccessKey {
		t.Fatal("expected credential presence flags")
	}

	if strings.Contains(sec.Settings["region"], fixtureAccessKey) {
		t.Fatal("secret leaked into settings")
	}
}

func TestParseCommentsColonDefaultInheritance(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	body := strings.Join([]string{
		"[DEFAULT]",
		"region: us-west-2",
		"; comment line",
		"[profile child]",
		"# inline # stays in value = not-applied",
		"output = json",
	}, "\n")

	writeConfig(t, dir, body, "")

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	child := cfg.Profiles["child"]
	if child.Settings["region"] != "us-west-2" {
		t.Fatalf("inherited region = %q", child.Settings["region"])
	}

	if child.Settings["output"] != "json" {
		t.Fatalf("output = %q", child.Settings["output"])
	}
}

func TestParseNestedBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	body := strings.Join([]string{
		"[profile nested]",
		"s3 =",
		"    signature_version = s3v4",
		"    addressing_style = path",
	}, "\n")

	writeConfig(t, dir, body, "")

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	nested := cfg.Profiles["nested"].Nested["s3"]
	if nested["signature_version"] != "s3v4" {
		t.Fatalf("nested s3: %+v", nested)
	}
}

func TestParseDuplicateKeyErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	body := "[profile dup]\nregion = a\nregion = b\n"
	writeConfig(t, dir, body, "")

	_, err := awsconfig.Load(env)
	if err == nil {
		t.Fatal("expected duplicate key error")
	}

	assertNoSecretInError(t, err, fixtureAccessKey)

	if !strings.Contains(err.Error(), "region") {
		t.Fatalf("error should name key: %v", err)
	}
}
