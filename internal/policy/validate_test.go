package policy_test

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/policy"
	"errors"
	"testing"
)

func TestLoadPresentDefaultsWithoutAction(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "defaults.action")
}

func TestLoadPresentCredentialExportWithoutAction(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[credential_export]
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "credential_export.action")
}

func TestLoadInvalidActionDoesNotEchoSecret(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[profiles.x]
action = "`+fixtureSecret+`"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "profiles.x.action")
	assertNoSecretInError(t, err)
}

func TestLoadInvalidApprovalTTL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		ttl  string
	}{
		{name: "zero", ttl: "0s"},
		{name: "negative", ttl: "-5m"},
		{name: "day", ttl: "1d"},
		{name: "bare", ttl: "5"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := configDirEnv(t)
			writePolicy(t, env, `version = 1

[profiles.x]
action = "prompt"
approval_ttl = "`+testCase.ttl+`"
`)

			_, err := policy.Load(env)
			if err == nil {
				t.Fatal("expected error")
			}

			assertPolicyInvalidKey(t, err, "profiles.x.approval_ttl")
		})
	}
}

func TestLoadAgentsSectionRejected(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[agents]
default = "deny"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindDecode {
		t.Fatalf("want KindDecode, got %v", err)
	}
}

func TestLoadUnknownProfileField(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[profiles.x]
action = "allow"
classification = "readonly"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindDecode {
		t.Fatalf("want KindDecode, got %v", err)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "")

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindDecode {
		t.Fatalf("want KindDecode, got %v", err)
	}
}

func TestLoadUnsupportedVersion(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 99\n")

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, decode.ErrUnsupportedVersion) {
		t.Fatalf("errors.Is ErrUnsupportedVersion = false, got %v", err)
	}
}

func TestLoadDuplicateKey(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1
version = 2
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, decode.ErrDuplicateKey) {
		t.Fatalf("errors.Is ErrDuplicateKey = false, got %v", err)
	}
}

func TestLoadProfileMissingAction(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[profiles.x]
approval_ttl = "5m"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "profiles.x.action")
}

func assertPolicyInvalidKey(t *testing.T, err error, key string) {
	t.Helper()

	var polErr *policy.Error
	if !errors.As(err, &polErr) {
		t.Fatalf("errors.As policy.Error = false, got %v", err)
	}

	if polErr.Kind != policy.KindInvalid {
		t.Fatalf("Kind = %v, want KindInvalid", polErr.Kind)
	}

	if polErr.Key != key {
		t.Fatalf("Key = %q, want %q", polErr.Key, key)
	}
}
