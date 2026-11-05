package policy

type document struct {
	Version          int                        `toml:"version"`
	Defaults         *defaultsDocument          `toml:"defaults"`
	Environment      *environmentDocument       `toml:"environment"`
	Agents           *agentsDocument            `toml:"agents"`
	Profiles         map[string]profileDocument `toml:"profiles"`
	CredentialExport *credentialExportDocument  `toml:"credential_export"`
}

type agentsDocument struct {
	Detection *agentsDetectionDocument `toml:"detection"`
}

type agentsDetectionDocument struct {
	EnvironmentMarkers *bool `toml:"environment_markers"`
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

type environmentDocument struct {
	Action *string `toml:"action"`
}
