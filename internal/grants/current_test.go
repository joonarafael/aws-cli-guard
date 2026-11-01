package grants_test

import (
	"aws-cli-guard/internal/grants"
	"testing"
)

func TestCurrent(t *testing.T) {
	t.Parallel()

	for _, testCase := range currentCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := grants.Current(testCase.list, testCase.profile, testCase.now, testCase.current)
			if got != testCase.want {
				t.Fatalf("Current() = %v, want %v", got, testCase.want)
			}
		})
	}
}
