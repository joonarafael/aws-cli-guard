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
	"time"
)

func TestRunAllowTTLRequiredMissingPolicy(t *testing.T) {
	t.Parallel()

	runAllowCancelWithoutConfirm(t, allowTestEnv(t), "nottl")
}

func TestRunAllowTTLRequiredNoApprovalTTL(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowPolicy(t, env, strings.Join([]string{
		allowPolicyVersion,
		"",
		allowPolicyDefault,
		allowPolicyPrompt,
		"",
		allowPolicyCredExp,
		allowPolicyPrompt,
		"",
		"[profiles.nottl]",
		allowPolicyPrompt,
	}, "\n"))

	runAllowCancelWithoutConfirm(t, env, "nottl")
}

func TestRunAllowCorruptPolicyWithoutTTL(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowPolicy(t, env, "broken\n")

	runAllowCancelWithoutConfirm(t, env, "nottl")
}

func TestRunAllowUnreadableAWSConfigBeforeConfirm(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowAWSConfig(t, env, "not valid aws config\n[[", "")

	called := false

	status := cli.Dispatch(
		allowArgvWithTTL(allowTestProfile),
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string, string) error {
			called = true

			return nil
		},
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	if called {
		t.Fatal("confirm should not run")
	}
}

func TestRunAllowStoresFingerprintWithoutSecret(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowAWSConfig(t, env, strings.Join([]string{
		"[profile fp]",
		"credential_process = " + allowCredProcTest,
	}, "\n"), "")

	var stdout, stderr bytes.Buffer

	status := cli.Dispatch(
		allowArgvWithTTL("fp"),
		env,
		&stdout,
		&stderr,
		matchPhrase("fp"),
	)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, stderr.String())
	}

	if strings.Contains(stdout.String(), allowCredProcTest) ||
		strings.Contains(stderr.String(), allowCredProcTest) {
		t.Fatal("credential_process command leaked")
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 1 || !got.Grants[0].Source.CredentialProcess {
		t.Fatalf("grants = %+v", got.Grants)
	}
}

func TestRunAllowReplaceGrantKeepsOtherProfile(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	first := time.Date(2027, 1, 1, 12, 0, 0, 0, time.UTC)
	other := time.Date(2027, 2, 1, 12, 0, 0, 0, time.UTC)

	oldSource := emptyAllowSource()
	oldSource.RoleARN = "arn:aws:iam::1:role/Old"

	err := grants.Replace(env, []grants.Grant{
		{Profile: "keep-me", Expires: other, Source: emptyAllowSource()},
		{Profile: allowTestProfile, Expires: first, Source: oldSource},
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	status := cli.Dispatch(
		allowArgvWithTTL(allowTestProfile),
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		matchPhrase(allowTestProfile),
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 2 {
		t.Fatalf("grants = %+v", got.Grants)
	}

	var kept, updated grants.Grant

	for _, grant := range got.Grants {
		switch grant.Profile {
		case "keep-me":
			kept = grant
		case allowTestProfile:
			updated = grant
		}
	}

	if !kept.Expires.Equal(other) {
		t.Fatalf("kept expires = %v", kept.Expires)
	}

	if updated.Expires.Equal(first) || updated.Source.RoleARN != "" {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestRunAllowCorruptGrantsUnchanged(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	stateDir := stateDirFromAllowEnv(t, env)
	grantsPath := filepath.Join(stateDir, "grants.json")

	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte("not-json\n")

	err = os.WriteFile(grantsPath, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	status := cli.Dispatch(
		allowArgvWithTTL(allowTestProfile),
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		matchPhrase(allowTestProfile),
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	got, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, body) {
		t.Fatal("corrupt grants file changed")
	}
}

func TestRunAllowUsageAndInvalidNoEcho(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "usage missing profile",
			args: []string{progName, allowSubcommand},
			want: "usage: aws-guard allow <profile> [--ttl <duration>]",
		},
		{
			name: "invalid ttl",
			args: []string{progName, allowSubcommand, allowTestProfile, allowTTLFlag, "1d"},
			want: "the ttl is invalid",
		},
		{
			name: "invalid profile control char",
			args: []string{progName, allowSubcommand, "bad\nname", allowTTLFlag, allowTTL5m},
			want: "the profile name is invalid",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := allowTestEnv(t)

			var stderr bytes.Buffer

			status := cli.Dispatch(testCase.args, env, &bytes.Buffer{}, &stderr, matchPhrase("x"))
			if status != output.CancelStatus {
				t.Fatalf("status = %d", status)
			}

			line := strings.TrimSuffix(stderr.String(), "\n")

			wantLine := "aws-cli-guard: error: " + testCase.want
			if line != wantLine {
				t.Fatalf("stderr = %q want %q", line, wantLine)
			}

			if strings.Contains(stderr.String(), "1d") ||
				strings.Contains(stderr.String(), "bad\nname") {
				t.Fatalf("stderr echoed argv: %q", stderr.String())
			}
		})
	}
}
