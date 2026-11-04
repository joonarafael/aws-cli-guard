package cli_test

import "strings"

const (
	grantsSubcommand     = "grants"
	grantsListSubcommand = "list"
	grantsListProfileA   = "alpha"
	grantsListProfileB   = "beta"
	grantsListNoGrants   = "aws-cli-guard: no grants\n"
	grantsListUnknownSub = "unknown"
)

func grantsListDispatchArgs() []string {
	return []string{progName, grantsSubcommand, grantsListSubcommand}
}

func stringsContainsUsageGrants(stderr string) bool {
	return strings.Contains(stderr, "usage: aws-guard grants list")
}

func matchingV2GrantJSON(profile string) string {
	return matchingV2GrantJSONWithRole(profile, "arn:aws:iam::123456789012:role/Admin")
}

func grantsListSortOrderJSON() string {
	return `{
  "version": 2,
  "grants": [
    {
      "profile": "zebra",
      "expires": "2099-06-01T00:00:00Z",
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
        "external_id": "",
        "chain": []
      }
    },
    {
      "profile": "beta",
      "expires": "2099-03-01T00:00:00Z",
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
        "external_id": "",
        "chain": []
      }
    },
    {
      "profile": "alpha",
      "expires": "2099-01-01T00:00:00Z",
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
        "external_id": "",
        "chain": []
      }
    }
  ]
}
`
}

func matchingV2GrantJSONWithRole(profile, roleARN string) string {
	return `{
  "version": 2,
  "grants": [
    {
      "profile": "` + profile + `",
      "expires": "2099-01-01T00:00:00Z",
      "source": {
        "account_id": "123456789012",
        "role_arn": "` + roleARN + `",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false,
        "source_profile": "",
        "credential_source": "",
        "sso_session": "",
        "sso_start_url": "",
        "sso_region": "",
        "external_id": "",
        "chain": []
      }
    }
  ]
}
`
}

func grantsListSSOProfJSON() string {
	return `{
  "version": 2,
  "grants": [
    {
      "profile": "sso-prof",
      "expires": "2099-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false,
        "source_profile": "",
        "credential_source": "",
        "sso_session": "",
        "sso_start_url": "https://example.com/sso",
        "sso_region": "us-east-1",
        "external_id": "",
        "chain": []
      }
    }
  ]
}
`
}

func grantsListCredProcJSON() string {
	return `{
  "version": 2,
  "grants": [
    {
      "profile": "proc",
      "expires": "2099-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": true,
        "source_profile": "",
        "credential_source": "",
        "sso_session": "",
        "sso_start_url": "",
        "sso_region": "",
        "external_id": "",
        "chain": []
      }
    }
  ]
}
`
}

func grantsListUnreadGrantJSON() string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "live",
      "expires": "2099-01-01T00:00:00Z",
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

func expiredRevokeGrantJSON(profile string) string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "` + profile + `",
      "expires": "2020-01-01T00:00:00Z",
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
