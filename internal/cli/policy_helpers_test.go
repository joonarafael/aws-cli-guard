package cli_test

import (
	"aws-cli-guard/internal/output"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	policySubcommand     = "policy"
	policyListSubcommand = "list"
	policyExplainSub     = "explain"
	configSubcommand     = "config"
	configValidateSub    = "validate"
	policyTestProfile    = "production-admin"
	policyTestProfileDev = "dev-readonly"
	policyExplainAccount = "123456789012"
	policyExplainRoleARN = "arn:aws:iam::123456789012:role/Admin"
	policyCredProcMarker = "UNIQUE-CRED-PROC-POLICY-EXPLAIN"
)

func policyTestEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	configDir := filepath.Join(home, "config")

	return []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_CONFIG_DIR=" + configDir,
	}
}

func policyTestEnvWithAWSConfig(t *testing.T) ([]string, string) {
	t.Helper()

	env := policyTestEnv(t)
	configPath := filepath.Join(t.TempDir(), "aws-config")

	return append(env, "AWS_CONFIG_FILE="+configPath), configPath
}

func writePolicyFile(t *testing.T, env []string, body string, mode os.FileMode) {
	t.Helper()

	dir := configDirFromPolicyEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.toml")

	if mode == 0 {
		mode = 0o600
	}

	err = os.WriteFile(path, []byte(body), mode)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func configDirFromPolicyEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "AWS_CLI_GUARD_CONFIG_DIR="); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_CONFIG_DIR missing")

	return ""
}

func assertPolicyCancel(t *testing.T, status int, stderr string) {
	t.Helper()

	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d; stderr = %q", status, output.CancelStatus, stderr)
	}
}

func assertStdoutContains(t *testing.T, stdout, needle string) {
	t.Helper()

	if !strings.Contains(stdout, needle) {
		t.Fatalf("stdout = %q, want substring %q", stdout, needle)
	}
}

func assertStdoutNotContains(t *testing.T, stdout, needle string) {
	t.Helper()

	if strings.Contains(stdout, needle) {
		t.Fatalf("stdout must not contain %q: %q", needle, stdout)
	}
}

func assertStderrContains(t *testing.T, stderr, needle string) {
	t.Helper()

	if !strings.Contains(stderr, needle) {
		t.Fatalf("stderr = %q, want substring %q", stderr, needle)
	}
}

func assertDetailOrder(t *testing.T, text string, prefixes ...string) {
	t.Helper()

	last := -1

	for _, prefix := range prefixes {
		idx := strings.Index(text, "aws-cli-guard: "+prefix)
		if idx < 0 {
			t.Fatalf("missing %q in %q", prefix, text)
		}

		if idx <= last {
			t.Fatalf("order wrong for %q in %q", prefix, text)
		}

		last = idx
	}
}
