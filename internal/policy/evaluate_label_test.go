package policy_test

import (
	"aws-cli-guard/internal/policy"
	"testing"
)

func TestBasisLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		basis policy.Basis
		want  string
	}{
		{policy.BasisProfile, "profile"},
		{policy.BasisDefault, "defaults"},
		{policy.BasisCredentialExport, "credential_export"},
		{policy.BasisEnvironment, "environment"},
		{policy.BasisAgentProfile, policy.BasisAgentProfile.Label()},
		{policy.BasisAgentDefault, policy.BasisAgentDefault.Label()},
		{policy.Basis(0), ""},
	}

	for _, testCase := range cases {
		got := testCase.basis.Label()
		if got != testCase.want {
			t.Fatalf("Basis(%d).Label() = %q, want %q", testCase.basis, got, testCase.want)
		}
	}
}
