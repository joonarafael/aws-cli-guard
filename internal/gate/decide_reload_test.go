package gate_test

import (
	"aws-cli-guard/internal/gate"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func reloadDecideArgs() []string {
	return []string{
		gateProgName,
		gateSTSService,
		gateGetCallerIdentityOp,
		gateProfileFlag,
		gateTestProfile,
	}
}

func grantsPathFromGateEnv(t *testing.T, env []string) string {
	t.Helper()

	return filepath.Join(stateDirFromEnv(t, env), "grants.json")
}

func TestDecideReloadDenyBeatsCurrentGrant(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))
	writeGateGrants(t, env, currentGrantBody(gateTestProfile))

	args := reloadDecideArgs()

	var stderr bytes.Buffer

	permitted := gate.Decide(args, env, &stderr)
	if !permitted {
		t.Fatalf("first Decide = false, stderr = %q", stderr.String())
	}

	grantsPath := grantsPathFromGateEnv(t, env)

	grantsBefore, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatalf("ReadFile grants: %v", err)
	}

	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))

	stderr.Reset()

	permitted = gate.Decide(args, env, &stderr)
	if permitted {
		t.Fatal("second Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "policy: deny")

	grantsAfter, err := os.ReadFile(grantsPath)
	if err != nil {
		t.Fatalf("ReadFile grants after: %v", err)
	}

	if string(grantsAfter) != string(grantsBefore) {
		t.Fatal("grants.json was modified")
	}
}

func TestDecideReloadAllowThenPromptRefuses(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	args := reloadDecideArgs()

	var stderr bytes.Buffer

	permitted := gate.Decide(args, env, &stderr)
	if !permitted {
		t.Fatalf("first Decide = false, stderr = %q", stderr.String())
	}

	writeGatePolicy(t, env, promptPolicyBody(gateTestProfile))

	stderr.Reset()

	permitted = gate.Decide(args, env, &stderr)
	if permitted {
		t.Fatal("second Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())
	assertContains(t, stderr.String(), "aws-guard allow "+gateTestProfile)

	grantsPath := grantsPathFromGateEnv(t, env)

	_, err := os.Stat(grantsPath)
	if err == nil {
		t.Fatal("grants.json must not be created")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Stat grants: %v", err)
	}
}

func TestDecideReloadCorruptPolicyRefuses(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	args := reloadDecideArgs()

	var stderr bytes.Buffer

	permitted := gate.Decide(args, env, &stderr)
	if !permitted {
		t.Fatalf("first Decide = false, stderr = %q", stderr.String())
	}

	corrupt := []byte("not valid toml\n")
	policyPath := policyPathFromEnv(t, env)

	err := os.WriteFile(policyPath, corrupt, 0o600)
	if err != nil {
		t.Fatalf("WriteFile policy: %v", err)
	}

	stderr.Reset()

	permitted = gate.Decide(args, env, &stderr)
	if permitted {
		t.Fatal("second Decide = true, want false")
	}

	assertContains(t, stderr.String(), "the policy file is unusable")

	after, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("ReadFile policy after: %v", err)
	}

	if string(after) != string(corrupt) {
		t.Fatal("corrupt policy file was modified")
	}

	grantsPath := grantsPathFromGateEnv(t, env)

	_, err = os.Stat(grantsPath)
	if err == nil {
		t.Fatal("grants.json must not be created")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Stat grants: %v", err)
	}
}
