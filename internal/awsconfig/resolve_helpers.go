package awsconfig

func emptyResolvedSource() ResolvedSource {
	return ResolvedSource{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
		SourceProfile:     "",
		CredentialSource:  "",
		SSOSession:        "",
		SSOStartURL:       "",
		SSORegion:         "",
		ExternalID:        "",
		Chain:             nil,
	}
}
