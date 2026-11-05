package policy_test

import (
	"aws-cli-guard/internal/policy"
	"errors"
	"testing"
)

const v4PolicyPrefix = `version = 4

[defaults]
action = "allow"

[environment]
action = "allow"

[credential_export]
action = "allow"
`

func TestLoadVersion4AgentActionsWithMarkersEnabled(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v4PolicyPrefix+`

[agents.detection]
environment_markers = true

[agents]
default = "deny"

[agents.profiles.dev-readonly]
action = "allow"
`)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !got.HasAgentDefault || got.AgentDefault != policy.ActionDeny {
		t.Fatalf("default: %+v", got)
	}

	if got.AgentProfiles["dev-readonly"] != policy.ActionAllow {
		t.Fatalf("profiles: %+v", got.AgentProfiles)
	}
}

func TestLoadVersion4RejectsAgentActionsWhenMarkersOff(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v4PolicyPrefix+`

[agents.detection]
environment_markers = false

[agents]
default = "deny"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindInvalid {
		t.Fatalf("want KindInvalid, got %v", err)
	}

	if polErr.Key != policyKeyAgentsDefault {
		t.Fatalf("Key = %q", polErr.Key)
	}
}

func TestLoadVersion4RejectsAgentActionsWithoutDetectionTable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v4PolicyPrefix+`

[agents.profiles.staging]
action = "deny"
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Key != "agents.profiles.staging" {
		t.Fatalf("got %v", err)
	}
}
