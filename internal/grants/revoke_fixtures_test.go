package grants_test

func revokeTwoProfileFixture() string {
	return `{
  "version": 1,
  "grants": [
    {
      "profile": "drop-me",
      "expires": "2000-01-01T00:00:00Z",
      "source": {
        "account_id": "",
        "role_arn": "",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    },
    {
      "profile": "keep-me",
      "expires": "2099-06-01T12:00:00Z",
      "source": {
        "account_id": "111122223333",
        "role_arn": "arn:aws:iam::111122223333:role/Keep",
        "sso_account_id": "",
        "sso_role_name": "",
        "credential_process": false
      }
    }
  ]
}
`
}
