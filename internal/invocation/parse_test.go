package invocation_test

import (
	"aws-cli-guard/internal/invocation"
	"strings"
	"testing"
)

const (
	awsBin          = "aws"
	profileOpt      = "--profile"
	profileNameProd = "prod"
	serviceEC2      = "ec2"
	opDescribe      = "describe-instances"
	secretQuery     = "SecretValueMustNotAppearInStruct"

	reasonAmbiguousGlobalOption    = "ambiguous-global-option"
	reasonGlobalOptionInvalidValue = "global-option-invalid-value"
)

func TestParseCommandOperationProfile(t *testing.T) {
	t.Parallel()
	runParseCases(t, commandOperationCases())
}

func TestParseGlobalOptionEdgeCases(t *testing.T) {
	t.Parallel()
	runParseCases(t, globalEdgeCases())
}

func TestParseFailureClearsFields(t *testing.T) {
	t.Parallel()
	runParseCases(t, failureClearCases())
}

func TestParseDoesNotRetainNonProfileValues(t *testing.T) {
	t.Parallel()

	args := []string{
		awsBin,
		"--query",
		secretQuery,
		"--endpoint-url",
		"https://secret.example",
		profileOpt,
		profileNameProd,
		serviceEC2,
		opDescribe,
	}

	got := invocation.Parse(args)
	if !got.OK {
		t.Fatalf("Parse() OK = false, reason %q", got.Reason)
	}

	if got.Profile != profileNameProd {
		t.Fatalf("Profile = %q, want %q", got.Profile, profileNameProd)
	}

	if got.Command != serviceEC2 || got.Operation != opDescribe {
		t.Fatalf("command/operation = %q / %q", got.Command, got.Operation)
	}

	blob := got.Command + got.Operation + got.Profile + got.Reason
	if strings.Contains(blob, secretQuery) || strings.Contains(blob, "secret.example") {
		t.Fatal("non-profile option values leaked into Invocation fields")
	}
}

type parseCase struct {
	name string
	args []string
	want invocation.Invocation
}

func runParseCases(t *testing.T, cases []parseCase) {
	t.Helper()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := invocation.Parse(testCase.args)
			if got != testCase.want {
				t.Fatalf("Parse() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func invOK(command, operation, profile string, hasProfile bool) invocation.Invocation {
	return invocation.Invocation{
		Command:    command,
		Operation:  operation,
		Profile:    profile,
		HasProfile: hasProfile,
		OK:         true,
		Reason:     "",
	}
}

func invFail(reason string) invocation.Invocation {
	return invocation.Invocation{
		Command:    "",
		Operation:  "",
		Profile:    "",
		HasProfile: false,
		OK:         false,
		Reason:     reason,
	}
}

func commandOperationCases() []parseCase {
	return []parseCase{
		{
			name: "nil args",
			args: nil,
			want: invOK("", "", "", false),
		},
		{
			name: "argv only",
			args: []string{awsBin},
			want: invOK("", "", "", false),
		},
		{
			name: "profile without command",
			args: []string{awsBin, profileOpt, profileNameProd},
			want: invOK("", "", profileNameProd, true),
		},
		{
			name: "profile after service",
			args: []string{awsBin, serviceEC2, opDescribe, profileOpt, profileNameProd},
			want: invOK(serviceEC2, opDescribe, profileNameProd, true),
		},
		{
			name: "configure export credentials",
			args: []string{awsBin, "configure", "export-credentials"},
			want: invOK("configure", "export-credentials", "", false),
		},
		{
			name: "unknown service param before profile",
			args: []string{
				awsBin,
				"s3api",
				"put-object",
				"--bucket",
				"mybucket",
				profileOpt,
				profileNameProd,
			},
			want: invOK("s3api", "put-object", profileNameProd, true),
		},
	}
}

func globalEdgeCases() []parseCase {
	return []parseCase{
		{
			name: "cli read timeout negative value",
			args: []string{
				awsBin,
				"--cli-read-timeout",
				"-1",
				profileOpt,
				profileNameProd,
				"s3",
				"ls",
			},
			want: invOK("s3", "ls", profileNameProd, true),
		},
		{
			name: "cli read timeout float",
			args: []string{awsBin, "--cli-read-timeout=1.5", "s3", "ls"},
			want: invFail(reasonGlobalOptionInvalidValue),
		},
		{
			name: "cli read timeout empty",
			args: []string{awsBin, "--cli-read-timeout=", "s3", "ls"},
			want: invFail(reasonGlobalOptionInvalidValue),
		},
		{
			name: "ambiguous no cli",
			args: []string{awsBin, "--no-cli", "s3", "ls"},
			want: invFail(reasonAmbiguousGlobalOption),
		},
		{
			name: "no cli pager abbreviation",
			args: []string{awsBin, "--no-cli-p", "s3", "ls"},
			want: invOK("s3", "ls", "", false),
		},
		{
			name: "single dash profile ignored",
			args: []string{awsBin, "-profile", profileNameProd, "s3", "ls"},
			want: invOK(profileNameProd, "s3", "", false),
		},
	}
}

func failureClearCases() []parseCase {
	reasons := []string{
		reasonAmbiguousGlobalOption,
		"global-option-unexpected-value",
		"global-option-missing-value",
		"profile-flag-missing-value",
		"empty-profile-name",
		reasonGlobalOptionInvalidValue,
	}
	cases := make([]parseCase, 0, len(reasons))

	for _, reason := range reasons {
		args := failureArgsForReason(reason)
		cases = append(cases, parseCase{
			name: reason,
			args: args,
			want: invFail(reason),
		})
	}

	return cases
}

func failureArgsForReason(reason string) []string {
	switch reason {
	case reasonAmbiguousGlobalOption:
		return []string{awsBin, "--no-cli", "s3", "ls"}
	case "global-option-unexpected-value":
		return []string{awsBin, "--debug=true", "s3", "ls"}
	case "global-option-missing-value":
		return []string{awsBin, "--region"}
	case "profile-flag-missing-value":
		return []string{awsBin, profileOpt}
	case "empty-profile-name":
		return []string{awsBin, "--profile=", "s3", "ls"}
	case reasonGlobalOptionInvalidValue:
		return []string{awsBin, "--output", "JSON", "s3", "ls"}
	default:
		return []string{awsBin, "s3", "ls"}
	}
}
