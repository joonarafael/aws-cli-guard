package grants

type document struct {
	Version int              `json:"version"`
	Grants  *[]grantDocument `json:"grants"`
}

type grantDocument struct {
	Profile *string         `json:"profile"`
	Expires *string         `json:"expires"`
	Source  *sourceDocument `json:"source"`
}

//nolint:tagliatelle // ADR 0011 snake_case JSON keys in grants.json.
type sourceDocument struct {
	AccountID         *string `json:"account_id"`
	RoleARN           *string `json:"role_arn"`
	SSOAccountID      *string `json:"sso_account_id"`
	SSORoleName       *string `json:"sso_role_name"`
	CredentialProcess *bool   `json:"credential_process"`
}

type encodeDocument struct {
	Version int           `json:"version"`
	Grants  []encodeGrant `json:"grants"`
}

type encodeGrant struct {
	Profile string       `json:"profile"`
	Expires string       `json:"expires"`
	Source  encodeSource `json:"source"`
}

//nolint:tagliatelle // ADR 0011 snake_case JSON keys in grants.json.
type encodeSource struct {
	AccountID         string `json:"account_id"`
	RoleARN           string `json:"role_arn"`
	SSOAccountID      string `json:"sso_account_id"`
	SSORoleName       string `json:"sso_role_name"`
	CredentialProcess bool   `json:"credential_process"`
}
