package cliversion_test

import (
	"aws-cli-guard/internal/cliversion"
	"testing"
)

func TestClassifyOutputSupported(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		stdout string
		want   string
	}{
		{"full", "aws-cli/2.9.0 Python/3.11.0 Linux/test exe/test", "2.9.0"},
		{"short", "aws-cli/2.9 Python/3.11.0", "2.9"},
		{"exact ceiling", "aws-cli/2.9.0", "2.9.0"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := cliversion.ClassifyOutput([]byte(testCase.stdout), nil)
			if got.Relation != cliversion.RelationSupported {
				t.Fatalf("relation = %v want supported", got.Relation)
			}

			if got.Version != testCase.want {
				t.Fatalf("version = %q want %q", got.Version, testCase.want)
			}
		})
	}
}

func TestClassifyOutputOlder(t *testing.T) {
	t.Parallel()

	got := cliversion.ClassifyOutput([]byte("aws-cli/2.8.9 foo"), nil)
	if got.Relation != cliversion.RelationOlder {
		t.Fatalf("relation = %v", got.Relation)
	}

	got = cliversion.ClassifyOutput([]byte("aws-cli/1.27.0"), nil)
	if got.Relation != cliversion.RelationOlder {
		t.Fatalf("v1 relation = %v", got.Relation)
	}
}

func TestClassifyOutputNewer(t *testing.T) {
	t.Parallel()

	got := cliversion.ClassifyOutput([]byte("aws-cli/2.9.1"), nil)
	if got.Relation != cliversion.RelationNewer {
		t.Fatalf("relation = %v", got.Relation)
	}

	got = cliversion.ClassifyOutput([]byte("aws-cli/2.9.0.1"), nil)
	if got.Relation != cliversion.RelationNewer {
		t.Fatalf("patch relation = %v", got.Relation)
	}
}

func TestClassifyOutputUnreadable(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"not a version",
		"aws-cli/2.9.0.dev0",
		"aws-cli/v2",
	}

	for _, raw := range cases {
		got := cliversion.ClassifyOutput([]byte(raw), nil)
		if got.Relation != cliversion.RelationUnreadable {
			t.Fatalf("input %q: relation = %v", raw, got.Relation)
		}
	}
}

func TestClassifyOutputStderrOnly(t *testing.T) {
	t.Parallel()

	got := cliversion.ClassifyOutput(nil, []byte("aws-cli/2.9.0\n"))
	if got.Relation != cliversion.RelationSupported {
		t.Fatalf("relation = %v", got.Relation)
	}
}

func TestPolicyAppliesAndWarns(t *testing.T) {
	t.Parallel()

	if cliversion.PolicyApplies(cliversion.RelationOlder) {
		t.Fatal("older should not apply policy")
	}

	if cliversion.PolicyApplies(cliversion.RelationUnreadable) {
		t.Fatal("unreadable should not apply policy")
	}

	if !cliversion.PolicyApplies(cliversion.RelationSupported) {
		t.Fatal("supported should apply policy")
	}

	if !cliversion.PolicyApplies(cliversion.RelationNewer) {
		t.Fatal("newer should apply policy")
	}

	if !cliversion.Warns(cliversion.RelationOlder) {
		t.Fatal("older should warn")
	}

	if cliversion.Warns(cliversion.RelationSupported) {
		t.Fatal("supported should not warn")
	}
}

func TestWarningText(t *testing.T) {
	t.Parallel()

	older := cliversion.WarningText(cliversion.Result{
		Relation: cliversion.RelationOlder,
		Version:  "2.8.9",
	})
	if older != "AWS CLI 2.8.9 is older than 2.9.0; running it with no policy" {
		t.Fatalf("older = %q", older)
	}

	unread := cliversion.WarningText(cliversion.Result{
		Relation: cliversion.RelationUnreadable,
		Version:  "",
	})
	if unread != "AWS CLI version could not be read; running it with no policy" {
		t.Fatalf("unread = %q", unread)
	}

	newer := cliversion.WarningText(cliversion.Result{
		Relation: cliversion.RelationNewer,
		Version:  "2.9.1",
	})

	wantNewer := "AWS CLI 2.9.1 is newer than the verified release 2.9.0; policy still applies"
	if newer != wantNewer {
		t.Fatalf("newer = %q", newer)
	}
}

func TestParseRelation(t *testing.T) {
	t.Parallel()

	rel, parsed := cliversion.ParseRelation("supported")
	if !parsed || rel != cliversion.RelationSupported {
		t.Fatalf("parse supported: parsed=%v rel=%v", parsed, rel)
	}

	_, parsed = cliversion.ParseRelation("bogus")
	if parsed {
		t.Fatal("expected false for bogus")
	}
}
