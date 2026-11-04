package grants_test

import (
	"aws-cli-guard/internal/grants"
	"errors"
	"time"
)

var errGrantStatusRead = errors.New("read failed")

type grantStatusCase struct {
	name    string
	grant   grants.Grant
	now     time.Time
	live    grants.Source
	readErr error
	want    grants.GrantStatusKind
}

type grantStatusFixture struct {
	expires time.Time
	before  time.Time
	grant   grants.Grant
}

func grantStatusFixtureData() grantStatusFixture {
	expires := time.Date(2026, 10, 28, 18, 0, 0, 0, time.UTC)
	source := currentMatchSource()

	return grantStatusFixture{
		expires: expires,
		before:  expires.Add(-time.Second),
		grant: grants.Grant{
			Profile: currentTestProfile,
			Expires: expires,
			Source:  source,
		},
	}
}

func grantStatusCases() []grantStatusCase {
	fix := grantStatusFixtureData()
	mismatch := currentMismatchSource()

	out := make([]grantStatusCase, 0, 7)
	out = append(out,
		grantStatusCase{
			name:    "one second before expiry current",
			grant:   fix.grant,
			now:     fix.before,
			live:    fix.grant.Source,
			readErr: nil,
			want:    grants.GrantStatusCurrent,
		},
		grantStatusCase{
			name:    "source mismatch",
			grant:   fix.grant,
			now:     fix.before,
			live:    mismatch,
			readErr: nil,
			want:    grants.GrantStatusSourceMismatch,
		},
		grantStatusCase{
			name:    "fingerprint error unread",
			grant:   fix.grant,
			now:     fix.before,
			live:    emptySource(),
			readErr: errGrantStatusRead,
			want:    grants.GrantStatusSourceUnread,
		},
	)

	out = append(out, grantStatusExpiryCases(fix)...)

	return out
}

func grantStatusExpiryCases(fix grantStatusFixture) []grantStatusCase {
	return []grantStatusCase{
		{
			name:    "exact expiry expired",
			grant:   fix.grant,
			now:     fix.expires,
			live:    fix.grant.Source,
			readErr: nil,
			want:    grants.GrantStatusExpired,
		},
		{
			name:    "after expiry expired",
			grant:   fix.grant,
			now:     fix.expires.Add(time.Second),
			live:    fix.grant.Source,
			readErr: nil,
			want:    grants.GrantStatusExpired,
		},
		{
			name: "empty sources current",
			grant: grants.Grant{
				Profile: currentTestProfile,
				Expires: fix.expires,
				Source:  emptySource(),
			},
			now:     fix.before,
			live:    emptySource(),
			readErr: nil,
			want:    grants.GrantStatusCurrent,
		},
		{
			name:    "expired wins over unread",
			grant:   fix.grant,
			now:     fix.expires,
			live:    emptySource(),
			readErr: errGrantStatusRead,
			want:    grants.GrantStatusExpired,
		},
	}
}
