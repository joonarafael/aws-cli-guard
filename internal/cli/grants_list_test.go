package cli_test

import (
	"aws-cli-guard/internal/cli"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGrantsListMissingFile(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)

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

	if stdout.String() != grantsListNoGrants {
		t.Fatalf("stdout = %q", stdout.String())
	}

	path := filepath.Join(revokeStateDir(t, env), "grants.json")

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatal("grants.json created")
	}
}

func TestRunGrantsListEmptyArray(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, `{"version": 2, "grants": []}`)

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

	if stdout.String() != grantsListNoGrants {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunGrantsListSortOrder(t *testing.T) {
	t.Parallel()

	env := revokeTestEnv(t)
	writeRevokeGrants(t, env, grantsListSortOrderJSON())

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

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: version: 2\n")
	assertDetailOrder(t, out,
		"profile: alpha",
		"profile: beta",
		"profile: zebra",
	)
}

func TestRunGrantsListCurrentWithIdentity(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowAWSConfig(t, env, strings.Join([]string{
		"[profile " + allowTestProfile + "]",
		"account_id = 123456789012",
		"role_arn = arn:aws:iam::123456789012:role/Admin",
	}, "\n"))
	writeRevokeGrants(t, env, matchingV2GrantJSON(allowTestProfile))

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

	out := stdout.String()
	assertStdoutContains(t, out, "aws-cli-guard: grant: current\n")
	assertStdoutContains(t, out, "aws-cli-guard: account: 123456789012\n")
}
