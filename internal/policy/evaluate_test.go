package policy_test

import (
	"aws-cli-guard/internal/policy"
	"testing"
)

func TestEvaluate(t *testing.T) {
	t.Parallel()

	for _, testCase := range evaluateCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := policy.Evaluate(testCase.pol, testCase.subj)
			if got != testCase.want {
				t.Fatalf("Evaluate() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func profileAction(action policy.Action) policy.Profile {
	return policy.Profile{
		Action:         action,
		ApprovalTTL:    0,
		HasApprovalTTL: false,
	}
}

func policyWithDefault(action policy.Action) policy.Policy {
	return policy.Policy{
		Path:                  "",
		Version:               0,
		Mode:                  0,
		DefaultAction:         action,
		EnvironmentAction:     policy.ActionPrompt,
		HasEnvironmentAction:  false,
		CredentialExport:      policy.ActionPrompt,
		EnvironmentMarkers:    false,
		HasEnvironmentMarkers: false,
		HasAgentDefault:       false,
		AgentDefault:          "",
		AgentProfiles:         map[string]policy.Action{},
		Profiles:              map[string]policy.Profile{},
	}
}

func policyWithProfiles(profiles map[string]policy.Profile) policy.Policy {
	return policy.Policy{
		Path:                  "",
		Version:               0,
		Mode:                  0,
		DefaultAction:         policy.ActionPrompt,
		EnvironmentAction:     policy.ActionPrompt,
		HasEnvironmentAction:  false,
		CredentialExport:      policy.ActionPrompt,
		EnvironmentMarkers:    false,
		HasEnvironmentMarkers: false,
		HasAgentDefault:       false,
		AgentDefault:          "",
		AgentProfiles:         map[string]policy.Action{},
		Profiles:              profiles,
	}
}
