package policy_test

import (
	"aws-cli-guard/internal/policy"
	"testing"
)

func TestEvaluateGrant(t *testing.T) {
	t.Parallel()

	for _, testCase := range evaluateGrantCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := policy.EvaluateGrant(
				testCase.decision,
				testCase.subj,
				testCase.list,
				testCase.now,
				testCase.current,
				testCase.readErr,
			)
			if got != testCase.want {
				t.Fatalf("EvaluateGrant() = %+v, want %+v", got, testCase.want)
			}

			if got.Permits() != testCase.wantPermits {
				t.Fatalf("Permits() = %v, want %v", got.Permits(), testCase.wantPermits)
			}
		})
	}
}
