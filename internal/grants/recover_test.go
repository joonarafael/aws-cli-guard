package grants_test

import (
	"aws-cli-guard/internal/grants"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecoverReplacesGarbage(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	path := grantsPathFromEnv(t, env)
	body := []byte(`{` + fixtureSecret + "\n")

	err := os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	aside, err := grants.Recover(env, time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 0 {
		t.Fatalf("grants = %+v", got.Grants)
	}

	asideData, err := os.ReadFile(aside)
	if err != nil {
		t.Fatalf("ReadFile aside: %v", err)
	}

	if !bytes.Equal(asideData, body) {
		t.Fatalf("aside bytes mismatch")
	}
}

func TestRecoverSymlinkLeavesTarget(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	dir := stateDirFromEnv(t, env)

	target := filepath.Join(dir, "real.json")
	targetBody := []byte(`{"` + fixtureSecret + `":true}`)

	err := os.WriteFile(target, targetBody, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	link := grantsPathFromEnv(t, env)

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatal(err)
	}

	_, err = grants.Recover(env, time.Now())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile target: %v", err)
	}

	if !bytes.Equal(got, targetBody) {
		t.Fatal("target unchanged")
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(file.Grants) != 0 {
		t.Fatalf("grants = %+v", file.Grants)
	}
}

func TestRecoverMissingCreatesNothing(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	_, err := grants.Recover(env, time.Now())
	if !errors.Is(err, grants.ErrMissing) {
		t.Fatalf("err = %v, want ErrMissing", err)
	}

	_, err = os.Stat(grantsPathFromEnv(t, env))
	if !os.IsNotExist(err) {
		t.Fatal("grants file should not be created")
	}
}

func TestPresentMissing(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	err := grants.Present(env)
	if !errors.Is(err, grants.ErrMissing) {
		t.Fatalf("err = %v, want ErrMissing", err)
	}
}
