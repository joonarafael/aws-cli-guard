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

type sourceDocument struct {
	AccountID         *string        `json:"account_id"`
	RoleARN           *string        `json:"role_arn"`
	SSOAccountID      *string        `json:"sso_account_id"`
	SSORoleName       *string        `json:"sso_role_name"`
	CredentialProcess *bool          `json:"credential_process"`
	SourceProfile     *string        `json:"source_profile"`
	CredentialSource  *string        `json:"credential_source"`
	SSOSession        *string        `json:"sso_session"`
	SSOStartURL       *string        `json:"sso_start_url"`
	SSORegion         *string        `json:"sso_region"`
	ExternalID        *string        `json:"external_id"`
	Chain             *[]hopDocument `json:"chain"`
}

type hopDocument struct {
	Profile           *string `json:"profile"`
	AccountID         *string `json:"account_id"`
	RoleARN           *string `json:"role_arn"`
	SSOAccountID      *string `json:"sso_account_id"`
	SSORoleName       *string `json:"sso_role_name"`
	CredentialProcess *bool   `json:"credential_process"`
	SourceProfile     *string `json:"source_profile"`
	CredentialSource  *string `json:"credential_source"`
	SSOSession        *string `json:"sso_session"`
	SSOStartURL       *string `json:"sso_start_url"`
	SSORegion         *string `json:"sso_region"`
	ExternalID        *string `json:"external_id"`
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

type encodeSource struct {
	AccountID         string      `json:"account_id"`
	RoleARN           string      `json:"role_arn"`
	SSOAccountID      string      `json:"sso_account_id"`
	SSORoleName       string      `json:"sso_role_name"`
	CredentialProcess bool        `json:"credential_process"`
	SourceProfile     string      `json:"source_profile"`
	CredentialSource  string      `json:"credential_source"`
	SSOSession        string      `json:"sso_session"`
	SSOStartURL       string      `json:"sso_start_url"`
	SSORegion         string      `json:"sso_region"`
	ExternalID        string      `json:"external_id"`
	Chain             []encodeHop `json:"chain"`
}

type encodeHop struct {
	Profile           string `json:"profile"`
	AccountID         string `json:"account_id"`
	RoleARN           string `json:"role_arn"`
	SSOAccountID      string `json:"sso_account_id"`
	SSORoleName       string `json:"sso_role_name"`
	CredentialProcess bool   `json:"credential_process"`
	SourceProfile     string `json:"source_profile"`
	CredentialSource  string `json:"credential_source"`
	SSOSession        string `json:"sso_session"`
	SSOStartURL       string `json:"sso_start_url"`
	SSORegion         string `json:"sso_region"`
	ExternalID        string `json:"external_id"`
}
