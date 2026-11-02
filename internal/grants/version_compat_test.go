package grants_test

import (
	"aws-cli-guard/internal/grants"
	"strings"
	"testing"
	"time"
)

func TestLoadVersion1GrantStillValid(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, sampleGrantJSON("production-admin", "2026-10-28T18:00:00Z"))

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 1 || len(got.Grants) != 1 {
		t.Fatalf("file: %+v", got)
	}

	if len(got.Grants[0].Source.Chain) != 0 {
		t.Fatalf("v1 chain: %+v", got.Grants[0].Source.Chain)
	}
}

func TestLoadVersion1RejectsSourceProfileKey(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	body := strings.Replace(sampleGrantJSON("p", "2026-10-28T18:00:00Z"),
		`"credential_process": false`,
		`"credential_process": false,
        "source_profile": ""`,
		1)

	writeGrants(t, env, body)

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestLoadVersion2MissingChainInvalid(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{
  "version": 2,
  "grants": [{
    "profile": "p",
    "expires": "2026-10-28T18:00:00Z",
    "source": {
      "account_id": "",
      "role_arn": "",
      "sso_account_id": "",
      "sso_role_name": "",
      "credential_process": false,
      "source_profile": "",
      "credential_source": "",
      "sso_session": "",
      "sso_start_url": "",
      "sso_region": "",
      "external_id": ""
    }
  }]
}
`)

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected invalid chain")
	}
}

func TestCurrentVersion1GrantMatchesV2Fingerprint(t *testing.T) {
	t.Parallel()

	stored := emptySource()
	stored.AccountID = "123456789012"
	stored.RoleARN = fixtureRoleARN

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		"[profile admin]",
		"role_arn = " + fixtureRoleARN,
		"account_id = 123456789012",
	}, "\n"), "")

	current, err := grants.Fingerprint(env, "admin")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	list := []grants.Grant{{Profile: "admin", Expires: now.Add(time.Hour), Source: stored}}

	if !grants.Current(list, "admin", now, current) {
		t.Fatal("expected v1 grant to match v2 fingerprint for plain role")
	}
}

func TestCurrentVersion1GrantFailsWhenChainAppears(t *testing.T) {
	t.Parallel()

	stored := emptySource()
	stored.RoleARN = fixtureRoleARN

	env := homeEnv(t)
	writeAWSConfig(t, env, strings.Join([]string{
		fixtureProfileBaseHeader,
		"role_arn = " + fixtureRoleARN,
		"",
		"[profile chained]",
		fixtureSourceProfileBase,
	}, "\n"), "")

	current, err := grants.Fingerprint(env, "chained")
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	list := []grants.Grant{{Profile: "chained", Expires: now.Add(time.Hour), Source: stored}}

	if grants.Current(list, "chained", now, current) {
		t.Fatal("expected v1 grant not to match chained profile")
	}
}

func TestUpdateRewritesFileVersion2PreservesOtherGrant(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, version1TwoGrantFixture())

	newSource := emptySource()
	newSource.RoleARN = fixtureRoleARN
	newSource.SourceProfile = fixtureProfileBaseName
	newSource.Chain = []grants.ChainHop{
		{
			Profile:           fixtureProfileBaseName,
			AccountID:         "",
			RoleARN:           fixtureRoleARN,
			SSOAccountID:      "",
			SSORoleName:       "",
			CredentialProcess: false,
			SourceProfile:     "",
			CredentialSource:  "",
			SSOSession:        "",
			SSOStartURL:       "",
			SSORegion:         "",
			ExternalID:        "",
		},
	}

	err := grants.Update(env, func(list []grants.Grant) ([]grants.Grant, error) {
		out := append([]grants.Grant(nil), list...)
		for i, g := range out {
			if g.Profile == "touch" {
				out[i].Source = newSource
			}
		}

		return out, nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 2 || len(got.Grants) != 2 {
		t.Fatalf("file: %+v", got)
	}

	if got.Grants[0].Source.RoleARN != "arn:aws:iam::111122223333:role/Keep" {
		t.Fatalf("keep grant changed: %+v", got.Grants[0].Source)
	}

	if len(got.Grants[0].Source.Chain) != 0 {
		t.Fatalf("keep grant chain rewritten: %+v", got.Grants[0].Source.Chain)
	}

	if got.Grants[1].Source.SourceProfile != fixtureProfileBaseName {
		t.Fatalf("touch grant: %+v", got.Grants[1].Source)
	}
}

func version1TwoGrantFixture() string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "keep",
      "expires": "2026-10-28T18:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "arn:aws:iam::111122223333:role/Keep",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    },
    {
      "profile": "touch",
      "expires": "2026-10-28T19:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    }
  ]
}
`
}
