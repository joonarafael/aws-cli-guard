package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/terminal"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunAllowWithTTLFlag(t *testing.T) {
	t.Parallel()

	runAllowSuccess(t, allowArgvWithTTL(allowTestProfile))
}

func TestRunAllowWithTTLEquals(t *testing.T) {
	t.Parallel()

	runAllowSuccess(t, []string{progName, allowSubcommand, allowTestProfile, "--ttl=" + allowTTL5m})
}

func TestRunAllowApprovalTTLFromPolicy(t *testing.T) {
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
		"[profiles.timed]",
		allowPolicyPrompt,
		`approval_ttl = "10m"`,
	}, "\n"))

	before := time.Now()

	status := cli.Dispatch(
		[]string{progName, allowSubcommand, "timed"},
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		matchPhrase("timed"),
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 1 {
		t.Fatalf("grants = %+v", got.Grants)
	}

	after := time.Now()
	wantMin := before.UTC().Add(10 * time.Minute).Truncate(time.Second)

	wantMax := after.UTC().Add(10 * time.Minute).Truncate(time.Second)
	if got.Grants[0].Expires.Before(wantMin) || got.Grants[0].Expires.After(wantMax) {
		t.Fatalf("expires = %v", got.Grants[0].Expires)
	}
}

func TestRunAllowDenyProfileStillWrites(t *testing.T) {
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
		"[profiles.blocked]",
		`action = "deny"`,
	}, "\n"))

	status := cli.Dispatch(
		allowArgvWithTTL("blocked"),
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		matchPhrase("blocked"),
	)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 1 || got.Grants[0].Profile != "blocked" {
		t.Fatalf("grants = %+v", got.Grants)
	}
}

func TestRunAllowExplicitTTLOverCorruptPolicy(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	writeAllowPolicy(t, env, "broken\n")

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

	loadSingleGrant(t, env)
}

func TestRunAllowConfirmMismatchLeavesFile(t *testing.T) {
	t.Parallel()

	env := allowTestEnv(t)
	stateDir := stateDirFromAllowEnv(t, env)
	grantsPath := filepath.Join(stateDir, "grants.json")

	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`{"version":1,"grants":[]}` + "\n")

	err = os.WriteFile(grantsPath, body, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	status := cli.Dispatch(
		allowArgvWithTTL(allowTestProfile),
		env,
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string, string) error { return terminal.ErrMismatch },
	)
	if status != output.CancelStatus {
		t.Fatalf("status = %d", status)
	}

	got, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, body) {
		t.Fatal("grants file changed on mismatch")
	}
}
