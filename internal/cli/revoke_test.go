package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/grants"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunRevokeProfileSuccess(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, sampleRevokeGrantJSON("prod"))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, "prod"},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if stdout.String() != "aws-cli-guard: revoked profile: prod\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Grants) != 0 {
		t.Fatalf("grants = %d", len(file.Grants))
	}
}

func TestRunRevokeAllSuccess(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	body := `{
  "version": 1,
  "grants": [
    {
      "profile": "a",
      "expires": "2099-01-01T00:00:00Z",
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
      "expires": "2000-01-01T00:00:00Z",
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
	writeRevokeGrants(t, env, body)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, revokeAllFlag},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if stdout.String() != "aws-cli-guard: revoked grants: 2\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRevokeNoGrantForProfile(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, "missing"},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	want := "aws-cli-guard: no grant for profile: missing\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}

	path := filepath.Join(revokeStateDir(t, env), "grants.json")

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatal("grants.json created")
	}
}

func TestRunRevokeAllNoGrants(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)

	var stdout bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, revokeAllFlag},
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	if stdout.String() != "aws-cli-guard: no grants\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
