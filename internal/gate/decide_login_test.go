package gate_test

import (
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/policy"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecideSSOLoginWithoutPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(gateSSOLoginArgs(), env, &stderr)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	grantsPath := filepath.Join(stateDirFromEnv(t, env), "grants.json")

	_, err := os.Stat(grantsPath)
	if err == nil {
		t.Fatal("grants.json must not be created for login handoff")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Stat grants: %v", err)
	}
}

func TestDecideSSOLoginDefaultProfileNoWarning(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		gateSSOLoginArgs(gateProfileFlag, "default"),
		env,
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	out := stderr.String()
	if strings.Contains(out, "aws-cli-guard: warning:") {
		t.Fatalf("stderr must not warn on default profile for login: %q", out)
	}

	if strings.Contains(out, "aws-cli-guard: error:") {
		t.Fatalf("stderr must not cancel login: %q", out)
	}
}

func TestDecideSSOLoginDenyPolicyStillPermits(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		gateSSOLoginArgs(gateProfileFlag, gateTestProfile),
		env,
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDecideSSOLoginCorruptFilesUnchanged(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, "not valid toml\n")
	writeGateGrants(t, env, "{not json")

	policyPath := policyPathFromEnv(t, env)
	grantsPath := filepath.Join(stateDirFromEnv(t, env), "grants.json")

	policyBefore, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("ReadFile policy: %v", err)
	}

	grantsBefore, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatalf("ReadFile grants: %v", err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(gateSSOLoginArgs(), env, &stderr)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	policyAfter, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("ReadFile policy after: %v", err)
	}

	if string(policyAfter) != string(policyBefore) {
		t.Fatal("corrupt policy file was modified")
	}

	grantsAfter, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatalf("ReadFile grants after: %v", err)
	}

	if string(grantsAfter) != string(grantsBefore) {
		t.Fatal("corrupt grants file was modified")
	}
}

func TestDecideEmptyProfileSSOLoginFailClosed(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "--profile=", gateSSOCommand, gateLoginOperation},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), policy.ReasonMissing)
}

func TestDecideTopLevelLoginMissingPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateLoginOperation},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), policy.ReasonMissing)
}

func TestDecideSSOLogoutMissingPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSSOCommand, gateLogoutOperation},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), policy.ReasonMissing)
}

func TestDecideSSOLogoutDenyPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateHandoffProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSSOCommand,
			gateLogoutOperation,
			gateProfileFlag,
			gateHandoffProfile,
		},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "policy: deny")
}
