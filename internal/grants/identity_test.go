package grants_test

import (
	"aws-cli-guard/internal/grants"
	"testing"
)

const (
	testAccountID   = "123456789012"
	testRoleARN     = "arn:aws:iam::123456789012:role/Admin"
	testSSOAccount  = "111122223333"
	testSSORoleName = "ReadOnly"
)

func TestIdentityLinesAccountAndRole(t *testing.T) {
	t.Parallel()

	source := identitySource()
	source.AccountID = testAccountID
	source.RoleARN = testRoleARN

	lines := grants.IdentityLines(source)
	want := []string{
		"account: " + testAccountID,
		"role: " + testRoleARN,
	}

	assertLinesEqual(t, lines, want)
}

func TestIdentityLinesSSOAccountOnly(t *testing.T) {
	t.Parallel()

	source := identitySource()
	source.SSOAccountID = testSSOAccount
	source.SSORoleName = testSSORoleName

	lines := grants.IdentityLines(source)
	want := []string{
		"account: " + testSSOAccount,
		"role: " + testSSORoleName,
	}

	assertLinesEqual(t, lines, want)
}

func TestIdentityLinesBothAccounts(t *testing.T) {
	t.Parallel()

	source := identitySource()
	source.AccountID = testAccountID
	source.SSOAccountID = testSSOAccount
	source.RoleARN = testRoleARN
	source.SSORoleName = testSSORoleName

	lines := grants.IdentityLines(source)
	want := []string{
		"account: " + testAccountID,
		"sso account: " + testSSOAccount,
		"role: " + testRoleARN,
		"sso role: " + testSSORoleName,
	}

	assertLinesEqual(t, lines, want)
}

func TestIdentityLinesCredentialProcess(t *testing.T) {
	t.Parallel()

	source := identitySource()
	source.CredentialProcess = true

	lines := grants.IdentityLines(source)

	const wantLine = "identity: configured source only; the caller was not verified"

	if len(lines) != 1 || lines[0] != wantLine {
		t.Fatalf("lines = %v", lines)
	}
}

func TestIdentityLinesEmpty(t *testing.T) {
	t.Parallel()

	lines := grants.IdentityLines(identitySource())
	if len(lines) != 0 {
		t.Fatalf("lines = %v, want empty", lines)
	}
}

func identitySource() grants.Source {
	return grants.Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
		SourceProfile:     "",
		CredentialSource:  "",
		SSOSession:        "",
		SSOStartURL:       "",
		SSORegion:         "",
		ExternalID:        "",
		Chain:             nil,
	}
}

func assertLinesEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("lines[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
