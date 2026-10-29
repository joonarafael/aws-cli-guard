package policy_test

import (
	"aws-cli-guard/internal/policy"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecoverReplacesGarbage(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "not-valid = [[[\n"+fixtureSecret+"\n")

	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	aside, err := policy.Recover(env, now)
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(policyDirFromEnv(t, env), "config.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatalf("policy body mismatch")
	}

	asideData, err := os.ReadFile(aside)
	if err != nil {
		t.Fatalf("ReadFile aside: %v", err)
	}

	if string(asideData) != "not-valid = [[[\n"+fixtureSecret+"\n" {
		t.Fatalf("aside should preserve old bytes")
	}
}

func TestRecoverSymlinkLeavesTarget(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	target := filepath.Join(dir, "real.toml")
	targetBody := []byte(fixtureSecret + "-target")

	err := os.WriteFile(target, targetBody, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "config.toml")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatal(err)
	}

	aside, err := policy.Recover(env, time.Now())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile target: %v", err)
	}

	if string(got) != string(targetBody) {
		t.Fatal("target unchanged")
	}

	info, err := os.Lstat(aside)
	if err != nil {
		t.Fatalf("Lstat aside: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("aside should be symlink")
	}

	data, err := os.ReadFile(link)
	if err != nil {
		t.Fatalf("ReadFile new policy: %v", err)
	}

	if string(data) != policy.DefaultFileBody {
		t.Fatal("new regular policy at live path")
	}
}

func TestRecoverMissingCreatesNothing(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)

	_, err := policy.Recover(env, time.Now())
	if !errors.Is(err, policy.ErrMissing) {
		t.Fatalf("err = %v, want ErrMissing", err)
	}

	_, err = os.Stat(filepath.Join(policyDirFromEnv(t, env), "config.toml"))
	if !os.IsNotExist(err) {
		t.Fatal("policy file should not be created")
	}
}
