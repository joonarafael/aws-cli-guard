package policy_test

import (
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"errors"
	"strings"
	"testing"
)

func TestLoadVersion1WithoutEnvironmentTable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = "allow"
`)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.HasEnvironmentAction {
		t.Fatal("HasEnvironmentAction = true, want false")
	}

	subj := subject.Subject{Kind: subject.Environment, Name: "", Reason: ""}

	dec := policy.Evaluate(got, subj)
	if dec.Action != policy.ActionAllow || dec.Basis != policy.BasisDefault {
		t.Fatalf("Evaluate() = %+v, want allow/default", dec)
	}
}

func TestLoadVersion1RejectsEnvironmentTable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[environment]
action = "deny"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "environment")
	assertNoSecretInError(t, err)
}

func TestLoadVersion2EnvironmentActions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		action string
		want   policy.Action
	}{
		{name: "allow", action: "allow", want: policy.ActionAllow},
		{name: "prompt", action: "prompt", want: policy.ActionPrompt},
		{name: "deny", action: "deny", want: policy.ActionDeny},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := configDirEnv(t)
			writePolicy(t, env, `version = 2

[defaults]
action = "`+testCase.action+`"

[environment]
action = "`+testCase.action+`"
`)

			got, err := policy.Load(env)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if !got.HasEnvironmentAction || got.EnvironmentAction != testCase.want {
				t.Fatalf(
					"environment: has=%v action=%q",
					got.HasEnvironmentAction,
					got.EnvironmentAction,
				)
			}

			subj := subject.Subject{Kind: subject.Environment, Name: "", Reason: ""}

			dec := policy.Evaluate(got, subj)
			if dec.Action != testCase.want || dec.Basis != policy.BasisEnvironment {
				t.Fatalf("Evaluate() = %+v", dec)
			}
		})
	}
}

func TestLoadVersion2OmittedEnvironmentUsesDefaults(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 2

[defaults]
action = "deny"
`)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.HasEnvironmentAction {
		t.Fatal("HasEnvironmentAction = true")
	}

	subj := subject.Subject{Kind: subject.Environment, Name: "", Reason: ""}

	dec := policy.Evaluate(got, subj)
	if dec.Action != policy.ActionDeny || dec.Basis != policy.BasisDefault {
		t.Fatalf("Evaluate() = %+v", dec)
	}
}

func TestEvaluateEnvironmentOverridesProfileAllow(t *testing.T) {
	t.Parallel()

	pol := policy.Policy{
		Path:                  "",
		Version:               2,
		Mode:                  0,
		DefaultAction:         policy.ActionPrompt,
		EnvironmentAction:     policy.ActionDeny,
		HasEnvironmentAction:  true,
		CredentialExport:      policy.ActionPrompt,
		EnvironmentMarkers:    false,
		HasEnvironmentMarkers: false,
		Profiles: map[string]policy.Profile{
			"prod": {Action: policy.ActionAllow, ApprovalTTL: 0, HasApprovalTTL: false},
		},
	}

	subj := subject.Subject{Kind: subject.Environment, Name: "", Reason: ""}

	dec := policy.Evaluate(pol, subj)
	if dec.Action != policy.ActionDeny || dec.Basis != policy.BasisEnvironment {
		t.Fatalf("Evaluate() = %+v", dec)
	}
}

func TestEvaluateUnresolvedIgnoresEnvironmentAction(t *testing.T) {
	t.Parallel()

	pol := policy.Policy{
		Path:                  "",
		Version:               2,
		Mode:                  0,
		DefaultAction:         policy.ActionPrompt,
		EnvironmentAction:     policy.ActionDeny,
		HasEnvironmentAction:  true,
		CredentialExport:      policy.ActionPrompt,
		EnvironmentMarkers:    false,
		HasEnvironmentMarkers: false,
		Profiles:              map[string]policy.Profile{},
	}

	subj := subject.Subject{Kind: subject.Unresolved, Name: "", Reason: "ambient"}

	dec := policy.Evaluate(pol, subj)
	if dec.Action != policy.ActionPrompt || dec.Basis != policy.BasisDefault {
		t.Fatalf("Evaluate() = %+v", dec)
	}
}

func TestLoadVersion2EmptyEnvironmentTable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 2

[environment]
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "environment.action")
}

func TestLoadVersion2InvalidEnvironmentAction(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[environment]
action = "`+fixtureSecret+`"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	assertPolicyInvalidKey(t, err, "environment")
	assertNoSecretInError(t, err)
}

func TestLoadVersion2UnknownEnvironmentField(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 2

[environment]
action = "prompt"
extra = true
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

func TestLoadVersion2AgentsStillRejected(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 2

[agents]
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindInvalid {
		t.Fatalf("want KindInvalid, got %v", err)
	}

	if polErr.Key != policyKeyAgents {
		t.Fatalf("Key = %q, want %q", polErr.Key, policyKeyAgents)
	}
}

func TestLoadVersion2FullRoundTrip(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	body := strings.TrimSpace(policy.DefaultFileBody)
	writePolicy(t, env, body)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 2 || !got.HasEnvironmentAction {
		t.Fatalf("policy: %+v", got)
	}
}
