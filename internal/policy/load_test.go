package policy_test

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFullPolicy(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = "prompt"

[profiles.dev-readonly]
action = "allow"

[profiles.production-admin]
action = "prompt"
approval_ttl = "5m"

[credential_export]
action = "deny"
`)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.DefaultAction != policy.ActionPrompt {
		t.Fatalf("DefaultAction = %q", got.DefaultAction)
	}

	if got.CredentialExport != policy.ActionDeny {
		t.Fatalf("CredentialExport = %q", got.CredentialExport)
	}

	prof := got.Profiles["production-admin"]
	if prof.Action != policy.ActionPrompt || !prof.HasApprovalTTL ||
		prof.ApprovalTTL != 5*time.Minute {
		t.Fatalf("production-admin: %+v", prof)
	}
}

func TestLoadOmittedDefaultsAndCredentialExport(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\n")

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.DefaultAction != policy.ActionPrompt || got.CredentialExport != policy.ActionPrompt {
		t.Fatalf("defaults: action=%q export=%q", got.DefaultAction, got.CredentialExport)
	}

	if len(got.Profiles) != 0 {
		t.Fatalf("profiles = %+v", got.Profiles)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, err = policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, policy.ErrMissing) || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is missing: got %v", err)
	}
}

func TestLoadLoosePermissionsStillLoads(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicyPerm(t, env, "version = 1\n", 0o644)

	_, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoadSymlinkRefused(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	target := filepath.Join(dir, "real.toml")

	err = os.WriteFile(target, []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	link := filepath.Join(dir, "config.toml")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	_, err = policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, files.ErrSymlink) {
		t.Fatalf("errors.Is ErrSymlink = false, got %v", err)
	}

	assertNoSecretInError(t, err)
}

func TestLoadDirectoryPathErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	configPath := filepath.Join(dir, "config.toml")

	err = os.Mkdir(configPath, 0o700)
	if err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	_, err = policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, files.ErrNotRegular) {
		t.Fatalf("errors.Is ErrNotRegular = false, got %v", err)
	}
}

func TestLoadHomeUnset(t *testing.T) {
	t.Parallel()

	_, err := policy.Load(nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("errors.Is ErrHomeUnset = false, got %v", err)
	}
}

func TestLoadTooLarge(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)

	body := make([]byte, (1<<20)+1)
	for i := range body {
		body[i] = ' '
	}

	copy(body, []byte("version = 1\n"))

	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.toml")

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, policy.ErrTooLarge) {
		t.Fatalf("errors.Is ErrTooLarge = false, got %v", err)
	}
}
