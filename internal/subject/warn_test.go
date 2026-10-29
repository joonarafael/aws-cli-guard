package subject_test

import (
	"aws-cli-guard/internal/subject"
	"bytes"
	"strings"
	"testing"
)

const wantDefaultWarningLine = "aws-cli-guard: warning: " +
	"using the default profile; a named profile is better practice"

func TestWarnDefaultProfileImplicitDefault(t *testing.T) {
	t.Parallel()
	assertDefaultWarning(t, []string{awsBin, "s3", "ls"}, nil)
}

func TestWarnDefaultProfileAWSProfileEnv(t *testing.T) {
	t.Parallel()
	assertDefaultWarning(
		t,
		[]string{awsBin, "s3", "ls"},
		[]string{"AWS_PROFILE=" + profileNameDefault},
	)
}

func TestWarnDefaultProfileExplicitFlag(t *testing.T) {
	t.Parallel()
	assertDefaultWarning(
		t,
		[]string{awsBin, profileOpt, profileNameDefault, "s3", "ls"},
		nil,
	)
}

func TestWarnDefaultProfileNoWarningForNamedProfile(t *testing.T) {
	t.Parallel()
	assertNoDefaultWarning(t, []string{awsBin, profileOpt, profileNameProd, "s3", "ls"}, nil)
}

func TestWarnDefaultProfileNoWarningForEnvironmentCredentials(t *testing.T) {
	t.Parallel()
	assertNoDefaultWarning(
		t,
		[]string{awsBin, "s3", "ls"},
		[]string{
			"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
			"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
		},
	)
}

func TestWarnDefaultProfileNoWarningForPartialAccessKey(t *testing.T) {
	t.Parallel()
	assertNoDefaultWarning(
		t,
		[]string{awsBin, "s3", "ls"},
		[]string{"AWS_ACCESS_KEY_ID=" + fixtureAccessKey},
	)
}

func TestWarnDefaultProfileNoWarningForAmbientCredentials(t *testing.T) {
	t.Parallel()
	assertNoDefaultWarning(
		t,
		[]string{awsBin, "s3", "ls"},
		[]string{"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI=/v2/creds"},
	)
}

func assertDefaultWarning(t *testing.T, args, env []string) {
	t.Helper()

	var buf bytes.Buffer

	subj := subject.Resolve(args, env)
	subject.WarnDefaultProfile(&buf, subj)

	line := strings.TrimSuffix(buf.String(), "\n")
	if line != wantDefaultWarningLine {
		t.Fatalf("stderr = %q want %q", line, wantDefaultWarningLine)
	}
}

func assertNoDefaultWarning(t *testing.T, args, env []string) {
	t.Helper()

	var buf bytes.Buffer

	subj := subject.Resolve(args, env)
	subject.WarnDefaultProfile(&buf, subj)

	if buf.Len() != 0 {
		t.Fatalf("expected no warning, got %q", buf.String())
	}
}
