package grants_test

import (
	"aws-cli-guard/internal/grants"
	"os"
	"path/filepath"
	"testing"
)

func TestRevokeMissingStateDirCreatesNothing(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	env := []string{
		"HOME=" + home,
		"AWS_CLI_GUARD_STATE_DIR=" + stateDir,
	}

	removed, err := grants.Revoke(env, grants.RevokeRequest{All: true, Profile: ""})
	if err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	if removed != 0 {
		t.Fatalf("removed = %d", removed)
	}

	_, err = os.Stat(stateDir)
	if !os.IsNotExist(err) {
		t.Fatalf("state dir created: %v", err)
	}
}

func TestRevokeMissingFileInExistingDirCreatesNothing(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	path := grantsPathFromEnv(t, env)

	removed, err := grants.Revoke(env, grants.RevokeRequest{All: false, Profile: "x"})
	if err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	if removed != 0 {
		t.Fatalf("removed = %d", removed)
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatal("grants.json created")
	}
}

func TestRevokeNoMatchLeavesBytes(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := sampleGrantJSON("keep-me", "2099-01-01T00:00:00Z")
	writeGrants(t, env, body)

	path := grantsPathFromEnv(t, env)

	removed, err := grants.Revoke(env, grants.RevokeRequest{All: false, Profile: "other"})
	if err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	if removed != 0 {
		t.Fatalf("removed = %d", removed)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != body {
		t.Fatal("grants file changed")
	}
}

func TestRevokeProfileRemovesOneKeepsOther(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	writeGrants(t, env, revokeTwoProfileFixture())

	removed, err := grants.Revoke(env, grants.RevokeRequest{All: false, Profile: "drop-me"})
	if err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	if removed != 1 {
		t.Fatalf("removed = %d", removed)
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(file.Grants) != 1 {
		t.Fatalf("grants len = %d", len(file.Grants))
	}

	if file.Grants[0].Profile != "keep-me" {
		t.Fatalf("profile = %q", file.Grants[0].Profile)
	}

	if file.Grants[0].Source.RoleARN != "arn:aws:iam::111122223333:role/Keep" {
		t.Fatalf("source = %+v", file.Grants[0].Source)
	}
}

func TestRevokeAllEmptiesFile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := `{
  "version": 1,
  "grants": [
    {
      "profile": "a",
      "expires": "2000-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    },
    {
      "profile": "b",
      "expires": "2099-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    }
  ]
}
`
	writeGrants(t, env, body)

	removed, err := grants.Revoke(env, grants.RevokeRequest{All: true, Profile: ""})
	if err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	if removed != 2 {
		t.Fatalf("removed = %d", removed)
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(file.Grants) != 0 {
		t.Fatalf("grants len = %d", len(file.Grants))
	}

	path := grantsPathFromEnv(t, env)

	_, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal("grants.json removed")
	}
}

func TestRevokeCorruptFileUnchanged(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := []byte("not-json\n")
	path := grantsPathFromEnv(t, env)
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = grants.Revoke(env, grants.RevokeRequest{All: true, Profile: ""})
	if err == nil {
		t.Fatal("expected error")
	}

	note, ok := grants.Corruption(err)
	if !ok {
		t.Fatalf("Corruption = false, err = %v", err)
	}

	if note.Class.Recover == "" {
		t.Fatal("missing recover hint")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(body) {
		t.Fatal("corrupt file changed")
	}
}
