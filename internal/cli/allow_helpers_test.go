package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/terminal"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	allowSubcommand   = "allow"
	allowTestProfile  = "production-admin"
	allowCredProcTest = "UNIQUE-CRED-PROC-COMMAND-ALLOW-TEST"
	allowTTLFlag      = "--ttl"
	allowTTL5m        = "5m"

	allowPolicyVersion = "version = 1"
	allowPolicyDefault = "[defaults]"
	allowPolicyPrompt  = `action = "prompt"`
	allowPolicyCredExp = "[credential_export]"
)

func allowArgvWithTTL(profile string) []string {
	return []string{progName, allowSubcommand, profile, allowTTLFlag, allowTTL5m}
}

func runAllowSuccess(t *testing.T, argv []string) {
	t.Helper()

	env := allowTestEnv(t)
	before := time.Now()

	var stdout, stderr bytes.Buffer

	status := cli.Dispatch(argv, env, &stdout, &stderr, matchPhrase(allowTestProfile))
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	after := time.Now()
	grant := loadSingleGrant(t, env)
	assertAllowGrantTiming(t, grant, before, after, 5*time.Minute)
	assertAllowStdout(t, stdout.String(), grant)
}

func runAllowCancelWithoutConfirm(t *testing.T, env []string, profile string) {
	t.Helper()

	called := false

	status := cli.Dispatch(
		[]string{progName, allowSubcommand, profile},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string, string) error {
			called = true

			return nil
		},
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if called {
		t.Fatal("confirm should not run")
	}

	_, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func loadSingleGrant(t *testing.T, env []string) grants.Grant {
	t.Helper()

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 1 {
		t.Fatalf("grants = %+v", got.Grants)
	}

	return got.Grants[0]
}

func assertAllowGrantTiming(
	t *testing.T,
	grant grants.Grant,
	before, after time.Time,
	ttl time.Duration,
) {
	t.Helper()

	if grant.Profile != allowTestProfile {
		t.Fatalf("profile = %q", grant.Profile)
	}

	if !grant.Source.Equal(emptyAllowSource()) {
		t.Fatalf("source = %+v", grant.Source)
	}

	wantMin := before.UTC().Add(ttl).Truncate(time.Second)

	wantMax := after.UTC().Add(ttl).Truncate(time.Second)
	if grant.Expires.Before(wantMin) || grant.Expires.After(wantMax) {
		t.Fatalf("expires = %v, want between %v and %v", grant.Expires, wantMin, wantMax)
	}
}

func assertAllowStdout(t *testing.T, out string, grant grants.Grant) {
	t.Helper()

	expLine := "expires: " + grant.Expires.Format(time.RFC3339)

	if !strings.Contains(out, "profile: "+allowTestProfile) {
		t.Fatalf("stdout = %q", out)
	}

	if !strings.Contains(out, expLine) {
		t.Fatalf("stdout = %q", out)
	}

	if !strings.Contains(out, "including headless ones") {
		t.Fatalf("stdout = %q", out)
	}
}

func emptyAllowSource() grants.Source {
	return grants.Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
	}
}

func allowTestEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	return []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_STATE_DIR=" + filepath.Join(home, "state"),
		"AWS_CLI_GUARD_CONFIG_DIR=" + filepath.Join(home, "config"),
	}
}

func stateDirFromAllowEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "AWS_CLI_GUARD_STATE_DIR="); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_STATE_DIR missing")

	return ""
}

func writeAllowPolicy(t *testing.T, env []string, body string) {
	t.Helper()

	dir := configDirFromAllowEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func configDirFromAllowEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "AWS_CLI_GUARD_CONFIG_DIR="); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_CONFIG_DIR missing")

	return ""
}

func writeAllowAWSConfig(t *testing.T, env []string, configBody, credBody string) {
	t.Helper()

	home := homeFromAllowEnv(t, env)
	awsDir := filepath.Join(home, ".aws")

	err := os.Mkdir(awsDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	if configBody != "" {
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configBody), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}

	if credBody != "" {
		err = os.WriteFile(filepath.Join(awsDir, "credentials"), []byte(credBody), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func homeFromAllowEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "HOME="); ok {
			return after
		}
	}

	t.Fatal("HOME missing")

	return ""
}

func matchPhrase(want string) func(string, string) error {
	return func(phrase, _ string) error {
		if phrase != want {
			return terminal.ErrMismatch
		}

		return nil
	}
}
