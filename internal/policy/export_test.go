package policy_test

import (
	"aws-cli-guard/internal/policy"
	"testing"
)

func TestEvaluateExport(t *testing.T) {
	t.Parallel()

	for _, testCase := range evaluateExportCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := policy.EvaluateExport(testCase.pol, testCase.printsAccessKeys)
			if got != testCase.want {
				t.Fatalf("EvaluateExport() = %+v, want %+v", got, testCase.want)
			}

			if got.Refuses() != testCase.wantRefuses {
				t.Fatalf("Refuses() = %v, want %v", got.Refuses(), testCase.wantRefuses)
			}
		})
	}
}

func TestReasonExportRefused(t *testing.T) {
	t.Parallel()

	const want = "commands that print access keys are refused until credential_export action is allow"
	if policy.ReasonExportRefused != want {
		t.Fatalf("ReasonExportRefused = %q, want %q", policy.ReasonExportRefused, want)
	}
}

type evaluateExportCase struct {
	name             string
	pol              policy.Policy
	printsAccessKeys bool
	want             policy.ExportDecision
	wantRefuses      bool
}

func evaluateExportCases() []evaluateExportCase {
	return append(
		append(evaluateExportNoApplyCases(), evaluateExportRefuseCases()...),
		evaluateExportAllowCases()...,
	)
}

func evaluateExportNoApplyCases() []evaluateExportCase {
	noApply := policy.ExportDecision{Applies: false, Action: "", Basis: policy.Basis(0)}

	return []evaluateExportCase{
		{
			name:             "rule does not apply",
			pol:              policyWithCredentialExport(policy.ActionDeny),
			printsAccessKeys: false,
			want:             noApply,
			wantRefuses:      false,
		},
	}
}

func evaluateExportRefuseCases() []evaluateExportCase {
	exportDeny := policy.ExportDecision{
		Applies: true,
		Action:  policy.ActionDeny,
		Basis:   policy.BasisCredentialExport,
	}

	return []evaluateExportCase{
		{
			name: "profile allow does not bypass export deny",
			pol: policyWithProfilesAndExport(map[string]policy.Profile{
				evalProfileDev: profileAction(policy.ActionAllow),
			}, policy.ActionDeny),
			printsAccessKeys: true,
			want:             exportDeny,
			wantRefuses:      true,
		},
		{
			name:             "export prompt refuses",
			pol:              profileAllowWithExport(policy.ActionPrompt),
			printsAccessKeys: true,
			want: policy.ExportDecision{
				Applies: true,
				Action:  policy.ActionPrompt,
				Basis:   policy.BasisCredentialExport,
			},
			wantRefuses: true,
		},
		{
			name:             "export deny refuses",
			pol:              profileAllowWithExport(policy.ActionDeny),
			printsAccessKeys: true,
			want:             exportDeny,
			wantRefuses:      true,
		},
		{
			name:             "unknown export action becomes deny",
			pol:              policyWithCredentialExport(policy.Action("maybe")),
			printsAccessKeys: true,
			want:             exportDeny,
			wantRefuses:      true,
		},
	}
}

func evaluateExportAllowCases() []evaluateExportCase {
	return []evaluateExportCase{
		{
			name:             "export allow applies",
			pol:              profileAllowWithExport(policy.ActionAllow),
			printsAccessKeys: true,
			want: policy.ExportDecision{
				Applies: true,
				Action:  policy.ActionAllow,
				Basis:   policy.BasisCredentialExport,
			},
			wantRefuses: false,
		},
	}
}

func policyWithCredentialExport(export policy.Action) policy.Policy {
	pol := policyWithDefault(policy.ActionPrompt)
	pol.CredentialExport = export

	return pol
}

func policyWithProfilesAndExport(
	profiles map[string]policy.Profile,
	export policy.Action,
) policy.Policy {
	pol := policyWithProfiles(profiles)
	pol.CredentialExport = export

	return pol
}

func profileAllowWithExport(export policy.Action) policy.Policy {
	return policyWithProfilesAndExport(map[string]policy.Profile{
		evalProfileDev: profileAction(policy.ActionAllow),
	}, export)
}
