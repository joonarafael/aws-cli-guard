package policy_test

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"testing"
	"time"
)

const (
	agentEvalProfile      = "dev-readonly"
	agentEvalOtherProfile = "other"
)

func TestEvaluateForCallerRecognizedUsesAgentProfile(t *testing.T) {
	t.Parallel()

	pol := policyWithProfiles(map[string]policy.Profile{
		agentEvalProfile: profileAction(policy.ActionAllow),
	})
	pol.AgentProfiles = map[string]policy.Action{
		agentEvalProfile: policy.ActionDeny,
	}

	subj := subject.Subject{Kind: subject.NamedProfile, Name: agentEvalProfile, Reason: ""}
	got := policy.EvaluateForCaller(pol, subj, true)

	if got.Action != policy.ActionDeny || got.Basis != policy.BasisAgentProfile {
		t.Fatalf("got %+v", got)
	}
}

func TestEvaluateForCallerRecognizedUsesAgentDefault(t *testing.T) {
	t.Parallel()

	pol := policyWithDefault(policy.ActionAllow)
	pol.HasAgentDefault = true
	pol.AgentDefault = policy.ActionDeny
	pol.Profiles = map[string]policy.Profile{
		agentEvalOtherProfile: profileAction(policy.ActionAllow),
	}

	subj := subject.Subject{Kind: subject.NamedProfile, Name: agentEvalOtherProfile, Reason: ""}
	got := policy.EvaluateForCaller(pol, subj, true)

	if got.Action != policy.ActionDeny || got.Basis != policy.BasisAgentDefault {
		t.Fatalf("got %+v", got)
	}
}

func TestEvaluateForCallerUnrecognizedUsesProfilePolicy(t *testing.T) {
	t.Parallel()

	pol := policyWithProfiles(map[string]policy.Profile{
		agentEvalProfile: profileAction(policy.ActionAllow),
	})
	pol.HasAgentDefault = true
	pol.AgentDefault = policy.ActionDeny

	subj := subject.Subject{Kind: subject.NamedProfile, Name: agentEvalProfile, Reason: ""}
	got := policy.EvaluateForCaller(pol, subj, false)

	if got.Action != policy.ActionAllow || got.Basis != policy.BasisProfile {
		t.Fatalf("got %+v", got)
	}
}

func TestEvaluateForCallerRecognizedEnvironmentIgnoresAgent(t *testing.T) {
	t.Parallel()

	pol := policyWithDefault(policy.ActionPrompt)
	pol.HasAgentDefault = true
	pol.AgentDefault = policy.ActionDeny
	pol.HasEnvironmentAction = true
	pol.EnvironmentAction = policy.ActionAllow

	subj := subject.Subject{Kind: subject.Environment, Name: "", Reason: ""}
	got := policy.EvaluateForCaller(pol, subj, true)

	if got.Action != policy.ActionAllow || got.Basis != policy.BasisEnvironment {
		t.Fatalf("got %+v", got)
	}
}

func TestEvaluateGrantAgentDenyBeatsCurrentGrant(t *testing.T) {
	t.Parallel()

	decision := policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisAgentProfile}
	subj := subject.Subject{Kind: subject.NamedProfile, Name: "p", Reason: ""}
	now := time.Now().UTC()
	source := emptyGrantSource()
	grantList := []grants.Grant{
		{
			Profile: "p",
			Expires: now.Add(time.Hour),
			Source:  source,
		},
	}

	got := policy.EvaluateGrant(decision, subj, grantList, now, source, nil)
	if got.Permits() {
		t.Fatal("agent deny must not permit")
	}
}
