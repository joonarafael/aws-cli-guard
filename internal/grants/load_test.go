package grants_test

import (
	"aws-cli-guard/internal/grants"
	"os"
	"testing"
	"time"
)

func TestLoadFullFile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("production-admin", "2026-10-28T18:00:00Z"))

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 1 || len(got.Grants) != 1 {
		t.Fatalf("file: %+v", got)
	}

	grant := got.Grants[0]
	if grant.Profile != "production-admin" {
		t.Fatalf("profile = %q", grant.Profile)
	}

	wantTime := time.Date(2026, 10, 28, 18, 0, 0, 0, time.UTC)
	if !grant.Expires.Equal(wantTime) {
		t.Fatalf("expires = %v", grant.Expires)
	}

	if grant.Source.RoleARN == "" || grant.Source.CredentialProcess {
		t.Fatalf("source: %+v", grant.Source)
	}
}

func TestLoadDottedProfileName(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("foo.bar", "2026-01-01T00:00:00Z"))

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Grants[0].Profile != "foo.bar" {
		t.Fatalf("profile = %q", got.Grants[0].Profile)
	}
}

func TestLoadOffsetExpiryTruncatesToUTC(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("p", "2026-06-15T12:30:45.999999999+02:00"))

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := time.Date(2026, 6, 15, 10, 30, 45, 0, time.UTC)
	if !got.Grants[0].Expires.Equal(want) {
		t.Fatalf("expires = %v want %v", got.Grants[0].Expires, want)
	}
}

func TestLoadExpiredGrantStillLoads(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("old", "2000-01-01T00:00:00Z"))

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Grants[0].Profile != "old" {
		t.Fatalf("profile = %q", got.Grants[0].Profile)
	}
}

func TestLoadMissingFileEmptyList(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Grants == nil {
		t.Fatal("Grants is nil")
	}

	if len(got.Grants) != 0 {
		t.Fatalf("len(Grants) = %d", len(got.Grants))
	}

	if got.Version != 1 {
		t.Fatalf("Version = %d", got.Version)
	}

	_, statErr := os.Stat(grantsPathFromEnv(t, env))
	if !os.IsNotExist(statErr) {
		t.Fatal("expected no grants file created")
	}
}

func TestLoadLoosePermissionsStillLoads(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrantsPerm(t, env, `{"version":1,"grants":[]}`+"\n", 0o644)

	_, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
}
