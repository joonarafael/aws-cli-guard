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

func TestDecideMissingSettingsAllowsWithPolicy(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDecideCorruptSettingsUnchanged(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))
	writeGateSettings(t, env, "not valid toml\n")

	path := filepath.Join(configDirFromEnv(t, env), "settings.toml")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "the settings file is unusable")
	assertContains(t, stderr.String(), "aws-guard recover settings")

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(after) != string(before) {
		t.Fatal("corrupt settings file was modified")
	}
}

func TestDecideVerifyTrueSettingsUnchanged(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))
	writeGateSettings(t, env, `version = 1

[identity]
verify = true
`)

	path := filepath.Join(configDirFromEnv(t, env), "settings.toml")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), "aws-guard recover settings")

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(after) != string(before) {
		t.Fatal("settings file was modified")
	}
}

func TestDecideLooseSettingsWarnsAndAllows(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))

	dir := configDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "settings.toml")

	err = os.WriteFile(path, []byte("version = 1\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	assertContains(t, stderr.String(), "aws-cli-guard: warning:")
	assertContains(t, stderr.String(), "settings file")
}

func TestDecideMissingPolicyIgnoresCorruptSettings(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGateSettings(t, env, "not valid toml\n")

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertContains(t, stderr.String(), policy.ReasonMissing)

	if strings.Contains(stderr.String(), "the settings file is unusable") {
		t.Fatalf("settings should not load when policy is missing: %q", stderr.String())
	}
}

func TestDecideSSOLoginIgnoresCorruptSettings(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGateSettings(t, env, "not valid toml\n")

	path := filepath.Join(configDirFromEnv(t, env), "settings.toml")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var stderr bytes.Buffer

	permitted := gate.Decide(gateSSOLoginArgs(), env, "", &stderr)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(after) != string(before) {
		t.Fatal("corrupt settings file was modified")
	}
}
