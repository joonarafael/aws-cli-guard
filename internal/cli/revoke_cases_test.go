package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRevokeCorruptGrants(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	dir := revokeStateDir(t, env)
	path := filepath.Join(dir, "grants.json")
	body := []byte("not-json\n")

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, revokeAllFlag},
		env,
		&bytes.Buffer{},
		&stderr,
		panicIfConfirm,
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	out := stderr.String()
	if !strings.Contains(out, "the grants file is unusable") {
		t.Fatalf("stderr = %q", out)
	}

	if !strings.Contains(out, "aws-guard recover grants") {
		t.Fatalf("stderr = %q", out)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, body) {
		t.Fatal("file changed")
	}
}

func TestRunRevokeProfileNamedAll(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, `{
  "version": 1,
  "grants": [
    {
      "profile": "all",
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
      "profile": "other",
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
`)

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, "all"},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Grants) != 1 || file.Grants[0].Profile != "other" {
		t.Fatalf("grants = %+v", file.Grants)
	}
}

func TestRunRevokeProfileWithSpaceAndDot(t *testing.T) {
	t.Parallel()

	const name = "my profile.v2"

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, sampleRevokeGrantJSON(name))

	status := cli.Dispatch(
		[]string{progName, revokeSubcommand, name},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	file, err := grants.Load(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Grants) != 0 {
		t.Fatalf("grants = %d", len(file.Grants))
	}
}

func TestRunRevokeUsageAndInvalidNoEcho(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "usage missing target",
			args: []string{progName, revokeSubcommand},
			want: "usage: aws-guard revoke <profile>|--all",
		},
		{
			name: "extra args",
			args: []string{progName, revokeSubcommand, "p", cliExtraArg},
			want: "usage: aws-guard revoke <profile>|--all",
		},
		{
			name: "invalid profile control char",
			args: []string{progName, revokeSubcommand, "bad\nname"},
			want: "the profile name is invalid",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := revokeTestEnv(t)

			var stderr bytes.Buffer

			status := cli.Dispatch(testCase.args, env, &bytes.Buffer{}, &stderr, panicIfConfirm)
			if status != output.CancelStatus {
				t.Fatalf("status = %d", status)
			}

			line := strings.TrimSuffix(stderr.String(), "\n")

			wantLine := "aws-cli-guard: error: " + testCase.want
			if line != wantLine {
				t.Fatalf("stderr = %q want %q", line, wantLine)
			}

			if strings.Contains(stderr.String(), "bad\nname") {
				t.Fatalf("stderr echoed argv")
			}
		})
	}
}
