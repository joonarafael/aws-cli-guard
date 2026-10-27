package subject_test

import (
	"aws-cli-guard/internal/subject"
	"strings"
	"testing"
)

func TestResolveProfileFlags(t *testing.T) {
	t.Parallel()
	runResolveCases(t, profileResolveCases())
}

func TestResolveCredentials(t *testing.T) {
	t.Parallel()
	runResolveCases(t, credentialResolveCases())
	runResolveCases(t, credentialEnvOrderCases())
}

func TestResolveAmbient(t *testing.T) {
	t.Parallel()
	runResolveCases(t, ambientResolveCases())
}

func TestResolveGlobalOptions(t *testing.T) {
	t.Parallel()
	runResolveCases(t, globalOptionResolveCases())
}

func runResolveCases(t *testing.T, cases []resolveCase) {
	t.Helper()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := subject.Resolve(testCase.args, testCase.env)
			if got != testCase.want {
				t.Fatalf("Resolve() = %+v, want %+v", got, testCase.want)
			}

			assertNoSecretLeak(t, got)
		})
	}
}

func assertNoSecretLeak(t *testing.T, sub subject.Subject) {
	t.Helper()

	secrets := []string{fixtureAccessKey, fixtureSecretKey, fixtureSession}
	blob := sub.Name + sub.Reason

	for _, secret := range secrets {
		if strings.Contains(blob, secret) {
			t.Fatalf("subject fields contain secret material")
		}
	}
}
