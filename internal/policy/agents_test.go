package policy_test

import (
	"aws-cli-guard/internal/policy"
	"errors"
	"fmt"
	"testing"
)

const v3PolicyPrefix = `version = 3

[defaults]
action = "allow"

[environment]
action = "prompt"

[credential_export]
action = "prompt"
`

func TestLoadVersion3WithoutAgentsTable(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 3 || got.HasEnvironmentMarkers || got.EnvironmentMarkers {
		t.Fatalf("policy: %+v", got)
	}
}

func TestLoadVersion3EnvironmentMarkersTrueAndFalse(t *testing.T) {
	t.Parallel()

	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprintf("environment_markers_%v", value), func(t *testing.T) {
			t.Parallel()

			env := configDirEnv(t)
			body := v3PolicyPrefix + `

[agents.detection]
environment_markers = ` + boolLiteral(value) + `
`
			writePolicy(t, env, body)

			got, err := policy.Load(env)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if !got.HasEnvironmentMarkers || got.EnvironmentMarkers != value {
				t.Fatalf("policy: %+v, want markers %v", got, value)
			}
		})
	}
}

func boolLiteral(v bool) string {
	if v {
		return "true"
	}

	return "false"
}

func TestLoadVersion3DetectionTableMissingEnvironmentMarkers(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`

[agents.detection]
`)

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	var polErr *policy.Error
	if !errors.As(err, &polErr) || polErr.Kind != policy.KindInvalid {
		t.Fatalf("want KindInvalid, got %v", err)
	}

	if polErr.Key != "agents.detection.environment_markers" {
		t.Fatalf("Key = %q", polErr.Key)
	}
}

func TestLoadVersion1AgentsDetectionRejected(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = "allow"

[agents.detection]
environment_markers = true
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

func TestLoadVersion3RejectsAgentsDefault(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`

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

func TestLoadVersion3RejectsParentProcess(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`

[agents.detection]
environment_markers = true
parent_process = false
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

func TestLoadVersion3RejectsAgentsProfiles(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`

[agents.profiles.dev-readonly]
action = "allow"
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

func TestLoadVersion3RejectsNonBooleanEnvironmentMarkers(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`

[agents.detection]
environment_markers = "yes"
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
