package policy_test

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"errors"
	"time"
)

const grantTestProfile = "production-admin"

var errGrantTestAWSConfigRead = errors.New("aws-cli-guard: grant test: aws config read failed")

type evaluateGrantCase struct {
	name        string
	decision    policy.Decision
	subj        subject.Subject
	list        []grants.Grant
	now         time.Time
	current     grants.Source
	readErr     error
	want        policy.GrantDecision
	wantPermits bool
}

func grantTestSource() grants.Source {
	return grants.Source{
		AccountID:         "123456789012",
		RoleARN:           "arn:aws:iam::123456789012:role/Admin",
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

func grantTestMismatchSource() grants.Source {
	return grants.Source{
		AccountID:         "",
		RoleARN:           "arn:aws:iam::999999999999:role/Other",
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

func emptyGrantSource() grants.Source {
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

func grantTestExpiry() time.Time {
	return time.Date(2026, 10, 28, 18, 0, 0, 0, time.UTC)
}

func evaluateGrantAllowDenyCases(
	allowDecision, denyDecision policy.Decision,
	grant grants.Grant,
	before time.Time,
	source grants.Source,
) []evaluateGrantCase {
	return []evaluateGrantCase{
		{
			name:        "allow empty list",
			decision:    allowDecision,
			subj:        namedProfile(grantTestProfile),
			list:        nil,
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionAllow, Current: false},
			wantPermits: true,
		},
		{
			name:        "deny beats matching grant",
			decision:    denyDecision,
			subj:        namedProfile(grantTestProfile),
			list:        []grants.Grant{grant},
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionDeny, Current: false},
			wantPermits: false,
		},
	}
}

func evaluateGrantPromptCases(
	promptDecision policy.Decision,
	grant grants.Grant,
	expires, before time.Time,
	source grants.Source,
) []evaluateGrantCase {
	return []evaluateGrantCase{
		{
			name:        "prompt with current grant",
			decision:    promptDecision,
			subj:        namedProfile(grantTestProfile),
			list:        []grants.Grant{grant},
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: true},
			wantPermits: true,
		},
		{
			name:        "prompt no grant",
			decision:    promptDecision,
			subj:        namedProfile(grantTestProfile),
			list:        nil,
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
		{
			name:        "prompt expired grant",
			decision:    promptDecision,
			subj:        namedProfile(grantTestProfile),
			list:        []grants.Grant{grant},
			now:         expires,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
		{
			name:        "prompt fingerprint mismatch",
			decision:    promptDecision,
			subj:        namedProfile(grantTestProfile),
			list:        []grants.Grant{grant},
			now:         before,
			current:     grantTestMismatchSource(),
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
	}
}

func evaluateGrantOtherProfileCase(
	promptDecision policy.Decision,
	expires, before time.Time,
	source grants.Source,
) evaluateGrantCase {
	return evaluateGrantCase{
		name:        "prompt grant other profile",
		decision:    promptDecision,
		subj:        namedProfile(grantTestProfile),
		list:        []grants.Grant{{Profile: "other", Expires: expires, Source: source}},
		now:         before,
		current:     source,
		readErr:     nil,
		want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
		wantPermits: false,
	}
}

func evaluateGrantSubjectCases(
	promptDecision policy.Decision,
	grant grants.Grant,
	before time.Time,
	source grants.Source,
) []evaluateGrantCase {
	return []evaluateGrantCase{
		{
			name:        "prompt environment subject",
			decision:    promptDecision,
			subj:        envSubject(""),
			list:        []grants.Grant{grant},
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
		{
			name:        "prompt unresolved subject",
			decision:    promptDecision,
			subj:        unresolvedSubject("", "ambient-credentials"),
			list:        []grants.Grant{grant},
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
		{
			name:        "prompt read error",
			decision:    promptDecision,
			subj:        namedProfile(grantTestProfile),
			list:        []grants.Grant{grant},
			now:         before,
			current:     emptyGrantSource(),
			readErr:     errGrantTestAWSConfigRead,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
		{
			name:        "prompt empty profile name",
			decision:    promptDecision,
			subj:        namedProfile(""),
			list:        []grants.Grant{grant},
			now:         before,
			current:     source,
			readErr:     nil,
			want:        policy.GrantDecision{Action: policy.ActionPrompt, Current: false},
			wantPermits: false,
		},
	}
}

func evaluateGrantCases() []evaluateGrantCase {
	expires := grantTestExpiry()
	before := expires.Add(-time.Minute)
	source := grantTestSource()
	grant := grants.Grant{Profile: grantTestProfile, Expires: expires, Source: source}
	promptDecision := policy.Decision{Action: policy.ActionPrompt, Basis: policy.BasisProfile}
	allowDecision := policy.Decision{Action: policy.ActionAllow, Basis: policy.BasisProfile}
	denyDecision := policy.Decision{Action: policy.ActionDeny, Basis: policy.BasisProfile}

	promptCases := evaluateGrantPromptCases(promptDecision, grant, expires, before, source)
	promptCases = append(
		promptCases,
		evaluateGrantOtherProfileCase(promptDecision, expires, before, source),
	)

	return append(
		append(
			evaluateGrantAllowDenyCases(allowDecision, denyDecision, grant, before, source),
			promptCases...,
		),
		evaluateGrantSubjectCases(promptDecision, grant, before, source)...,
	)
}
