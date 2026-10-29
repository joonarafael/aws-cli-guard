package policy_test

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/policy"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWritesDefaultPolicy(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)

	created, err := policy.Create(env)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !created {
		t.Fatal("expected created true")
	}

	path := filepath.Join(policyDirFromEnv(t, env), "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatalf("body mismatch")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 0600", info.Mode().Perm())
	}
}

func TestCreateLeavesExistingPolicy(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\n")

	created, err := policy.Create(env)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if created {
		t.Fatal("expected created false")
	}

	got, err := os.ReadFile(filepath.Join(policyDirFromEnv(t, env), "config.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != "version = 1\n" {
		t.Fatalf("policy was modified: %q", got)
	}
}

func TestCreateRefusesPolicySymlink(t *testing.T) {
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
		t.Fatalf("WriteFile: %v", err)
	}

	link := filepath.Join(dir, "config.toml")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	_, err = policy.Create(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error

	if !errors.As(err, &polErr) || polErr.Kind != policy.KindWrite {
		t.Fatalf("expected KindWrite, got %v", err)
	}

	if !errors.Is(err, files.ErrSymlink) {
		t.Fatalf("expected ErrSymlink, got %v", err)
	}
}
