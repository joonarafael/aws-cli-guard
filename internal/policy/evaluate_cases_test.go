package policy_test

import (
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"time"
)

const (
	evalProfileDev     = "dev"
	evalProfileProd    = "prod"
	evalProfileStaging = "staging"
)

type evaluateCase struct {
	name string
	pol  policy.Policy
	subj subject.Subject
	want policy.Decision
}

func evaluateProfileActionCases() []evaluateCase {
	return []evaluateCase{
		{
			name: "profile allow",
			pol: policyWithProfiles(map[string]policy.Profile{
				evalProfileDev: profileAction(policy.ActionAllow),
			}),
			subj: namedProfile(evalProfileDev),
			want: policy.Decision{Action: policy.ActionAllow, Basis: policy.BasisProfile},
		},
		{
			name: "profile prompt",
			pol: policyWithProfiles(map[string]policy.Profile{
				evalProfileStaging: profileAction(policy.ActionPrompt),
			}),
			subj: namedProfile(evalProfileStaging),
			want: policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisProfile},
		},
		{
			name: "profile deny",
			pol: policyWithProfiles(map[string]policy.Profile{
				evalProfileProd: profileAction(policy.ActionDeny),
			}),
			subj: namedProfile(evalProfileProd),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisProfile},
		},
		{
			name: "profile named default",
			pol: policyWithProfiles(map[string]policy.Profile{
				"default": profileAction(policy.ActionAllow),
			}),
			subj: namedProfile("default"),
			want: policy.Decision{Action: policy.ActionAllow, Basis: policy.BasisProfile},
		},
	}
}

func evaluateDefaultFallbackCases() []evaluateCase {
	return []evaluateCase{
		{
			name: "unspecified profile uses default allow",
			pol:  policyWithDefault(policy.ActionAllow),
			subj: namedProfile("missing"),
			want: policy.Decision{Action: policy.ActionAllow, Basis: policy.BasisDefault},
		},
		{
			name: "unspecified profile uses default prompt",
			pol:  policyWithDefault(policy.ActionPrompt),
			subj: namedProfile("missing"),
			want: policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisDefault},
		},
		{
			name: "unspecified profile uses default deny",
			pol:  policyWithDefault(policy.ActionDeny),
			subj: namedProfile("missing"),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisDefault},
		},
		{
			name: "nil profiles map uses default",
			pol: policy.Policy{
				Path:             "",
				Version:          0,
				Mode:             0,
				DefaultAction:    policy.ActionDeny,
				CredentialExport: policy.ActionPrompt,
				Profiles:         nil,
			},
			subj: namedProfile("any"),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisDefault},
		},
	}
}

func evaluateNonProfileSubjectCases() []evaluateCase {
	return []evaluateCase{
		{
			name: "environment uses default not profile allow",
			pol: policyWithProfiles(map[string]policy.Profile{
				evalProfileProd: profileAction(policy.ActionAllow),
			}),
			subj: envSubject(evalProfileProd),
			want: policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisDefault},
		},
		{
			name: "unresolved uses default",
			pol: policyWithProfiles(map[string]policy.Profile{
				evalProfileProd: profileAction(policy.ActionAllow),
			}),
			subj: unresolvedSubject(evalProfileProd, "ambient"),
			want: policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisDefault},
		},
	}
}

func evaluateScopeEdgeCases() []evaluateCase {
	return []evaluateCase{
		{
			name: "credential export deny does not change profile allow",
			pol: policy.Policy{
				Path:             "",
				Version:          0,
				Mode:             0,
				DefaultAction:    policy.ActionPrompt,
				CredentialExport: policy.ActionDeny,
				Profiles: map[string]policy.Profile{
					evalProfileDev: profileAction(policy.ActionAllow),
				},
			},
			subj: namedProfile(evalProfileDev),
			want: policy.Decision{Action: policy.ActionAllow, Basis: policy.BasisProfile},
		},
		{
			name: "approval ttl ignored",
			pol: policyWithProfiles(map[string]policy.Profile{
				"x": {
					Action:         policy.ActionPrompt,
					ApprovalTTL:    5 * time.Minute,
					HasApprovalTTL: true,
				},
			}),
			subj: namedProfile("x"),
			want: policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisProfile},
		},
		{
			name: "invalid profile action denies with profile basis",
			pol: policyWithProfiles(map[string]policy.Profile{
				"bad": profileAction(policy.Action("maybe")),
			}),
			subj: namedProfile("bad"),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisProfile},
		},
	}
}

func evaluateFailClosedCases() []evaluateCase {
	return []evaluateCase{
		{
			name: "empty default on miss denies",
			pol: policy.Policy{
				Path:             "",
				Version:          0,
				Mode:             0,
				DefaultAction:    "",
				CredentialExport: policy.ActionPrompt,
				Profiles:         map[string]policy.Profile{},
			},
			subj: namedProfile("none"),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisDefault},
		},
		{
			name: "empty default for environment denies",
			pol: policy.Policy{
				Path:             "",
				Version:          0,
				Mode:             0,
				DefaultAction:    "",
				CredentialExport: policy.ActionPrompt,
				Profiles:         map[string]policy.Profile{},
			},
			subj: envSubject(""),
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisDefault},
		},
		{
			name: "zero subject kind denies with zero basis",
			pol:  policyWithDefault(policy.ActionAllow),
			subj: subject.Subject{Kind: subject.Kind(0), Name: "", Reason: ""},
			want: policy.Decision{Action: policy.ActionDeny, Basis: policy.Basis(0)},
		},
	}
}

func evaluateCases() []evaluateCase {
	profile := evaluateProfileActionCases()
	fallback := evaluateDefaultFallbackCases()
	nonProfile := evaluateNonProfileSubjectCases()
	scope := evaluateScopeEdgeCases()
	failClosed := evaluateFailClosedCases()
	capacity := len(profile) + len(fallback) + len(nonProfile) + len(scope) + len(failClosed)
	out := make([]evaluateCase, 0, capacity)
	out = append(out, profile...)
	out = append(out, fallback...)
	out = append(out, nonProfile...)
	out = append(out, scope...)
	out = append(out, failClosed...)

	return out
}
