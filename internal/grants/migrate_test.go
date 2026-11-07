package grants_test

import (
	"aws-cli-guard/internal/grants"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMigrateVersion1PreservesSource(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	expires := "2026-12-01T12:00:00Z"
	writeGrants(t, env, sampleGrantJSON("admin", expires))

	result, err := grants.Migrate(env, time.Now())
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if !result.Rewrote || result.FromVersion != 1 || result.Aside == "" {
		t.Fatalf("result: %+v", result)
	}

	asideData, err := os.ReadFile(result.Aside)
	if err != nil {
		t.Fatalf("ReadFile aside: %v", err)
	}

	if !strings.Contains(string(asideData), `"version": 1`) {
		t.Fatal("aside should contain original version 1 bytes")
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 2 || len(got.Grants) != 1 {
		t.Fatalf("file: %+v", got)
	}

	grant := got.Grants[0]
	if grant.Profile != "admin" || grant.Source.RoleARN != "arn:aws:iam::123456789012:role/Admin" {
		t.Fatalf("grant: %+v", grant)
	}

	if len(grant.Source.Chain) != 0 {
		t.Fatal("expected empty chain without re-fingerprint")
	}
}

func TestMigrateVersion2NoOp(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := `{
  "version": 2,
  "grants": []
}
`
	writeGrants(t, env, body)

	result, err := grants.Migrate(env, time.Now())
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if result.Rewrote {
		t.Fatalf("result: %+v", result)
	}
}

func TestMigrateMissingFile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	_, err := grants.Migrate(env, time.Now())
	if err == nil || !errors.Is(err, grants.ErrMissing) {
		t.Fatalf("err = %v", err)
	}
}

func TestMigrateVersion1EncodesV2Keys(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("x", "2026-01-01T00:00:00Z"))

	_, err := grants.Migrate(env, time.Now())
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	raw, err := os.ReadFile(grantsPathFromEnv(t, env))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var doc struct {
		Version int `json:"version"`
	}

	err = json.Unmarshal(raw, &doc)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if doc.Version != 2 {
		t.Fatalf("version = %d", doc.Version)
	}

	if !strings.Contains(string(raw), `"chain"`) {
		t.Fatal("expected chain key in v2 output")
	}
}
