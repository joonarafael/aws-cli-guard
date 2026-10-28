package policy

type document struct {
	Version          int                        `toml:"version"`
	Defaults         *defaultsDocument          `toml:"defaults"`
	Profiles         map[string]profileDocument `toml:"profiles"`
	CredentialExport *credentialExportDocument  `toml:"credential_export"`
}

type defaultsDocument struct {
	Action *string `toml:"action"`
}

type profileDocument struct {
	Action      *string `toml:"action"`
	ApprovalTTL *string `toml:"approval_ttl"`
}

type credentialExportDocument struct {
	Action *string `toml:"action"`
}
