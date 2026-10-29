package policy_test

import (
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"os"
	"testing"
)

func TestPresentMissing(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	err = policy.Present(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, policy.ErrMissing) || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is missing: got %v", err)
	}
}

func TestPresentExists(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\n")

	err := policy.Present(env)
	if err != nil {
		t.Fatalf("Present: %v", err)
	}
}

func TestPresentHomeUnset(t *testing.T) {
	t.Parallel()

	err := policy.Present([]string{})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("expected ErrHomeUnset, got %v", err)
	}
}
