package cli_test

import (
	"aws-cli-guard/internal/cli"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const migrateSubcommand = "migrate"

func migrateTestEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	return []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_CONFIG_DIR=" + filepath.Join(home, "config"),
		"AWS_CLI_GUARD_STATE_DIR=" + filepath.Join(home, "state"),
	}
}

func writeGrantsFile(t *testing.T, env []string, body string, mode os.FileMode) {
	t.Helper()

	dir := stateDirFromMigrateEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "grants.json"), []byte(body), mode)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func stateDirFromMigrateEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_STATE_DIR="
		if after, ok := strings.CutPrefix(entry, prefix); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_STATE_DIR missing")

	return ""
}

func policyDirFromMigrateEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		const prefix = "AWS_CLI_GUARD_CONFIG_DIR="
		if after, ok := strings.CutPrefix(entry, prefix); ok {
			return after
		}
	}

	t.Fatal("AWS_CLI_GUARD_CONFIG_DIR missing")

	return ""
}

func TestRunMigrateMissingPolicy(t *testing.T) {
	t.Parallel()

	env := migrateTestEnv(t)

	var stdout, stderr bytes.Buffer

	code := cli.Dispatch([]string{progName, migrateSubcommand}, env, &stdout, &stderr, nil)
	if code != 3 {
		t.Fatalf("code = %d", code)
	}

	if !strings.Contains(stderr.String(), "missing") {
		t.Fatalf("stderr: %q", stderr.String())
	}
}

func TestRunMigrateBothFilesOld(t *testing.T) {
	t.Parallel()

	env := migrateTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"
`, 0o600)

	writeGrantsFile(t, env, `{
  "version": 1,
  "grants": []
}
`, 0o600)

	var stdout, stderr bytes.Buffer

	code := cli.Dispatch([]string{progName, migrateSubcommand}, env, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("code = %d stderr %q", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "policy: rewrote version 1 to version 4") {
		t.Fatalf("stdout: %q", out)
	}

	if !strings.Contains(out, "grants: rewrote version 1 to version 2") {
		t.Fatalf("stdout: %q", out)
	}
}

func TestRunMigratePolicyOnlyNoGrants(t *testing.T) {
	t.Parallel()

	env := migrateTestEnv(t)
	writePolicyFile(t, env, `version = 2

[defaults]
action = "prompt"

[environment]
action = "prompt"
`, 0o600)

	var stdout, stderr bytes.Buffer

	code := cli.Dispatch([]string{progName, migrateSubcommand}, env, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("code = %d", code)
	}

	if !strings.Contains(stdout.String(), "grants: no file to migrate") {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestRunMigrateExtraArg(t *testing.T) {
	t.Parallel()

	env := migrateTestEnv(t)
	writePolicyFile(t, env, "version = 4\n\n[defaults]\naction = \"prompt\"\n", 0o600)

	var stderr bytes.Buffer

	code := cli.Dispatch(
		[]string{progName, migrateSubcommand, cliExtraArg},
		env,
		&bytes.Buffer{},
		&stderr,
		nil,
	)
	if code != 3 {
		t.Fatalf("code = %d", code)
	}

	if !strings.Contains(stderr.String(), "usage: aws-guard migrate") {
		t.Fatalf("stderr: %q", stderr.String())
	}
}

func TestRunMigrateUnusableGrantsWritesNothing(t *testing.T) {
	t.Parallel()

	env := migrateTestEnv(t)
	writePolicyFile(t, env, `version = 1

[defaults]
action = "prompt"
`, 0o600)

	writeGrantsFile(t, env, `{"version": 99}`, 0o600)

	before := mustReadPolicyFile(t, env)

	var stderr bytes.Buffer

	code := cli.Dispatch([]string{progName, migrateSubcommand}, env, &bytes.Buffer{}, &stderr, nil)
	if code != 3 {
		t.Fatalf("code = %d", code)
	}

	after := mustReadPolicyFile(t, env)
	if after != before {
		t.Fatal("policy should be unchanged when grants preflight fails")
	}
}

func mustReadPolicyFile(t *testing.T, env []string) string {
	t.Helper()

	dir := policyDirFromMigrateEnv(t, env)

	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	return string(data)
}
