package grants_test

import (
	"aws-cli-guard/internal/grants"
	"testing"
)

func TestGrantStatusLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		kind grants.GrantStatusKind
		want string
	}{
		{grants.GrantStatusCurrent, "current"},
		{grants.GrantStatusExpired, "expired"},
		{grants.GrantStatusSourceMismatch, "source mismatch"},
		{grants.GrantStatusSourceUnread, "source unread"},
	}

	for _, testCase := range cases {
		if got := testCase.kind.Label(); got != testCase.want {
			t.Fatalf("Label() = %q, want %q", got, testCase.want)
		}
	}
}

func TestGrantStatus(t *testing.T) {
	t.Parallel()

	for _, testCase := range grantStatusCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := grants.GrantStatus(testCase.grant, testCase.now, testCase.live, testCase.readErr)
			if got != testCase.want {
				t.Fatalf("GrantStatus() = %v, want %v", got, testCase.want)
			}
		})
	}
}
