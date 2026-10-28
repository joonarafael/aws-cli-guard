package grants_test

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/paths"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOmittedGrantsInvalid(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{"version":1}`+"\n")

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	grErr, ok := errors.AsType[*grants.Error](err)
	if !ok || grErr.Kind != grants.KindInvalid {
		t.Fatalf("got %v", err)
	}
}

func TestLoadNullGrantsInvalid(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{"version":1,"grants":null}`+"\n")

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadDuplicateProfile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := `{
  "version": 1,
  "grants": [
    {
      "profile": "a",
      "expires": "2026-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    },
    {
      "profile": "a",
      "expires": "2026-01-02T00:00:00Z",
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

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	grErr, ok := errors.AsType[*grants.Error](err)
	if !ok || grErr.Key != "grants.1.profile" {
		t.Fatalf("got %v", err)
	}
}

func TestLoadControlCharacterInProfile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("bad\x07name", "2026-01-01T00:00:00Z"))

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadBadTimestampSecretSafe(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	secret := fixtureSecret + "-expires"
	writeGrants(t, env, sampleGrantJSON("p", secret))

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertNoSecretInError(t, err)
}

func TestLoadCredentialProcessStringType(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	secret := fixtureSecret + "-credproc"
	body := `{
  "version": 1,
  "grants": [{
    "profile": "p",
    "expires": "2026-01-01T00:00:00Z",
    "source": {
      "account_id": "",
      "role_arn": "",
      "sso_account_id": "",
      "sso_role_name": "",
      "credential_process": "` + secret + `"
    }
  }]
}
`
	writeGrants(t, env, body)

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertNoSecretInError(t, err)
}

func TestLoadUnknownField(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{"version":1,"grants":[],"extra":true}`+"\n")

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	grErr, ok := errors.AsType[*grants.Error](err)
	if !ok || grErr.Kind != grants.KindDecode {
		t.Fatalf("got %v", err)
	}
}

func TestLoadSymlinkRefused(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	target := filepath.Join(dir, "real.json")

	err = os.WriteFile(target, []byte(`{"version":1,"grants":[]}`), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	link := filepath.Join(dir, "grants.json")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	_, err = grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, files.ErrSymlink) {
		t.Fatalf("got %v", err)
	}
}

func TestLoadDirectoryPathErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_STATE_DIR=" + dir}

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	grantsPath := filepath.Join(dir, "grants.json")

	err = os.Mkdir(grantsPath, 0o700)
	if err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	_, err = grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, files.ErrNotRegular) {
		t.Fatalf("got %v", err)
	}
}

func TestLoadHomeUnset(t *testing.T) {
	t.Parallel()

	_, err := grants.Load(nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("got %v", err)
	}
}

func TestLoadTooLarge(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	body := make([]byte, (1<<20)+1)
	for i := range body {
		body[i] = ' '
	}

	copy(body, []byte(`{"version":1,"grants":[]}`))

	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "grants.json")

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, grants.ErrTooLarge) {
		t.Fatalf("got %v", err)
	}
}
