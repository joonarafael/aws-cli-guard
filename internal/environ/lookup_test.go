package environ_test

import (
	"aws-cli-guard/internal/environ"
	"testing"
)

func TestLookup(t *testing.T) {
	t.Parallel()

	for _, testCase := range lookupCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotVal, gotOK := environ.Lookup(testCase.key, testCase.env)
			if gotVal != testCase.wantVal || gotOK != testCase.wantOK {
				t.Fatalf(
					"Lookup(%q) = (%q, %v), want (%q, %v)",
					testCase.key,
					gotVal,
					gotOK,
					testCase.wantVal,
					testCase.wantOK,
				)
			}
		})
	}
}
