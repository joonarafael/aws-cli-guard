package awsconfig

// Config is the merged view of the AWS config and credentials files.
type Config struct {
	ConfigPath      string
	CredentialsPath string
	Profiles        map[string]Section
	SSOSessions     map[string]Section
	Services        map[string]Section
}

// Section holds non-secret settings for one profile, SSO session, or service block.
type Section struct {
	Settings             map[string]string
	Nested               map[string]map[string]string
	HasAccessKeyID       bool
	HasSecretAccessKey   bool
	HasSessionToken      bool
	HasCredentialProcess bool
}

func emptyConfig(configPath, credentialsPath string) Config {
	return Config{
		ConfigPath:      configPath,
		CredentialsPath: credentialsPath,
		Profiles:        map[string]Section{},
		SSOSessions:     map[string]Section{},
		Services:        map[string]Section{},
	}
}

func emptySection() Section {
	return Section{
		Settings:             map[string]string{},
		Nested:               map[string]map[string]string{},
		HasAccessKeyID:       false,
		HasSecretAccessKey:   false,
		HasSessionToken:      false,
		HasCredentialProcess: false,
	}
}
