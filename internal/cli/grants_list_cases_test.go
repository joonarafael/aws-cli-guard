package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGrantsListStoredSourceOmitsSSOURL(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, grantsListSSOProfJSON())

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutNotContains(t, stdout.String(), "example.com")
}

func TestRunGrantsListCredentialProcessIdentityLine(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, grantsListCredProcJSON())

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(),
		"identity: configured source only; the caller was not verified")
}

func TestRunGrantsListSourceMismatchMissingProfile(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowAWSConfig(t, env, strings.Join([]string{
		"[profile " + allowTestProfile + "]",
		"account_id = 123456789012",
		"role_arn = arn:aws:iam::123456789012:role/Admin",
	}, "\n"))

	const mismatchRole = "arn:aws:iam::999999999999:role/Other"

	writeRevokeGrants(t, env, matchingV2GrantJSONWithRole(allowTestProfile, mismatchRole))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: grant: source mismatch\n")
}

func TestRunGrantsListExpired(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, expiredRevokeGrantJSON("prod"))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: grant: expired\n")
}

func TestRunGrantsListExpiredDespiteUnreadConfig(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, expiredRevokeGrantJSON("prod"))
	env = append(env, "AWS_CONFIG_FILE="+revokeStateDir(t, env))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: grant: expired\n")
}

func TestRunGrantsListSourceUnread(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, grantsListUnreadGrantJSON())
	env = append(env, "AWS_CONFIG_FILE="+revokeStateDir(t, env))

	var stdout bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&bytes.Buffer{},
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStdoutContains(t, stdout.String(), "aws-cli-guard: grant: source unread\n")
}

func TestRunGrantsListLooseModeWarns(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	dir := revokeStateDir(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(dir, "grants.json"),
		[]byte(sampleRevokeGrantJSON("prod")),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	status := cli.Dispatch(
		grantsListDispatchArgs(),
		env,
		&stdout,
		&stderr,
		panicIfConfirm,
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	assertStderrContains(t, stderr.String(), "aws-cli-guard: warning:")
	assertStdoutContains(t, stdout.String(), "aws-cli-guard: profile: prod\n")
}

func TestRunGrantsListCorruptGrants(t *testing.T) {
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
		grantsListDispatchArgs(),
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

	if string(got) != string(body) {
		t.Fatal("grants file modified")
	}
}

func TestRunGrantsListUsage(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{progName, grantsSubcommand},
		{progName, grantsSubcommand, grantsListUnknownSub},
		{progName, grantsSubcommand, grantsListSubcommand, cliExtraArg},
	}

	for _, argv := range cases {
		var stderr bytes.Buffer

		status := cli.Dispatch(argv, nil, &bytes.Buffer{}, &stderr, panicIfConfirm)
		assertPolicyCancel(t, status, stderr.String())

		if !stringsContainsUsageGrants(stderr.String()) {
			t.Fatalf("argv %v stderr = %q", argv, stderr.String())
		}

		if strings.Contains(stderr.String(), grantsListUnknownSub) {
			t.Fatalf("stderr must not echo argv: %q", stderr.String())
		}
	}
}
