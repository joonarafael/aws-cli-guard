package gate_test

import (
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/policy"
	"bytes"
	"testing"
)

func TestDecideMissingGrantsMeansEmptyList(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateHandoffProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "s3", "ls", gateProfileFlag, gateHandoffProfile},
		env,
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}
}

func TestDecideCorruptGrants(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateHandoffProfile))
	writeGateGrants(t, env, "{not json")

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, "s3", "ls", gateProfileFlag, gateHandoffProfile},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "the grants file is unusable")
}

func TestDecideExportAllowStillProfilePolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateConfigureCommand,
			"get",
			"region",
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "policy: deny")
}

func TestDecidePathUnsetOnPresentOnlyBeforeLoad(t *testing.T) {
	t.Parallel()

	env := []string{"AWS_CLI_GUARD_CONFIG_DIR=" + t.TempDir()}

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		env,
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), policy.ReasonMissing)
}
