package grants_test

import (
	"aws-cli-guard/internal/grants"
	"time"
)

const currentTestProfile = "prod"

type currentCase struct {
	name    string
	list    []grants.Grant
	profile string
	now     time.Time
	current grants.Source
	want    bool
}

func currentMatchSource() grants.Source {
	out := emptySource()
	out.AccountID = "123456789012"
	out.RoleARN = fixtureRoleARN

	return out
}

func currentMismatchSource() grants.Source {
	out := emptySource()
	out.RoleARN = "arn:aws:iam::999999999999:role/Other"

	return out
}

func currentPartialMismatchSource() grants.Source {
	out := emptySource()
	out.RoleARN = "mismatch"

	return out
}

func currentExpiry() time.Time {
	return time.Date(2026, 10, 28, 18, 0, 0, 0, time.UTC)
}

func currentBeforeExpiryCases(expires time.Time, source grants.Source) []currentCase {
	before := expires.Add(-time.Second)

	return []currentCase{
		{
			name: "match one second before expiry",
			list: []grants.Grant{{
				Profile: currentTestProfile,
				Expires: expires,
				Source:  source,
			}},
			profile: currentTestProfile,
			now:     before,
			current: source,
			want:    true,
		},
		{
			name: "both sources empty",
			list: []grants.Grant{{
				Profile: currentTestProfile,
				Expires: expires,
				Source:  emptySource(),
			}},
			profile: currentTestProfile,
			now:     before,
			current: emptySource(),
			want:    true,
		},
		{
			name: "second grant matches",
			list: []grants.Grant{
				{
					Profile: currentTestProfile,
					Expires: expires,
					Source:  currentPartialMismatchSource(),
				},
				{Profile: currentTestProfile, Expires: expires, Source: source},
			},
			profile: currentTestProfile,
			now:     before,
			current: source,
			want:    true,
		},
	}
}

func currentRejectExpiryCases(expires time.Time, source grants.Source) []currentCase {
	return []currentCase{
		{
			name: "exact expiry not current",
			list: []grants.Grant{{
				Profile: currentTestProfile,
				Expires: expires,
				Source:  source,
			}},
			profile: currentTestProfile,
			now:     expires,
			current: source,
			want:    false,
		},
		{
			name: "after expiry",
			list: []grants.Grant{{
				Profile: currentTestProfile,
				Expires: expires,
				Source:  source,
			}},
			profile: currentTestProfile,
			now:     expires.Add(time.Second),
			current: source,
			want:    false,
		},
	}
}

func currentRejectOtherCases(expires time.Time, source grants.Source) []currentCase {
	before := expires.Add(-time.Second)

	return []currentCase{
		{
			name: "other profile",
			list: []grants.Grant{{
				Profile: "staging",
				Expires: expires,
				Source:  source,
			}},
			profile: currentTestProfile,
			now:     before,
			current: source,
			want:    false,
		},
		{
			name: "fingerprint mismatch",
			list: []grants.Grant{{
				Profile: currentTestProfile,
				Expires: expires,
				Source:  source,
			}},
			profile: currentTestProfile,
			now:     before,
			current: currentMismatchSource(),
			want:    false,
		},
		{
			name: "empty profile",
			list: []grants.Grant{
				{Profile: currentTestProfile, Expires: expires, Source: source},
			},
			profile: "",
			now:     before,
			current: source,
			want:    false,
		},
		{
			name:    "nil list",
			list:    nil,
			profile: currentTestProfile,
			now:     before,
			current: source,
			want:    false,
		},
	}
}

func currentRejectCases(expires time.Time, source grants.Source) []currentCase {
	return append(
		currentRejectExpiryCases(expires, source),
		currentRejectOtherCases(expires, source)...,
	)
}

func currentCases() []currentCase {
	expires := currentExpiry()
	source := currentMatchSource()

	return append(
		currentBeforeExpiryCases(expires, source),
		currentRejectCases(expires, source)...,
	)
}
