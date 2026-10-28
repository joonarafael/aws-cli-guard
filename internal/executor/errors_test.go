package executor_test

import (
	"aws-cli-guard/internal/executor"
	"errors"
	"strings"
	"testing"
)

func TestCancellationReasonSentinels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want string
	}{
		{executor.ErrRealAWSNotFound, "the real AWS CLI was not found"},
		{executor.ErrResolveWrapper, "the wrapper executable could not be resolved"},
		{executor.ErrStartFailed, "the real AWS CLI could not be started"},
		{
			executor.ErrPlatformUnsupported,
			"process replacement is not supported on this platform",
		},
	}

	for _, tc := range cases {
		got := executor.CancellationReason(tc.err)
		if got != tc.want {
			t.Errorf("CancellationReason(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

func TestCancellationReasonGenericOmitsUnderlyingText(t *testing.T) {
	t.Parallel()

	secret := "secret-path-marker-xyz"
	//nolint:err113 // deliberate non-sentinel to test the generic reason path
	got := executor.CancellationReason(errors.New(secret))
	want := "the command could not be started"

	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	if strings.Contains(got, secret) {
		t.Fatalf("reason must not contain underlying error text")
	}
}
