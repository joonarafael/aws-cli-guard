package grants

const identityUnverifiedLine = "identity: configured source only; the caller was not verified"

// IdentityLines returns account, role, and identity detail lines for a credential source.
// Lines omit empty values and never include SSO URLs, hop profiles, or credential-process commands.
func IdentityLines(source Source) []string {
	var lines []string

	lines = appendAccountLines(lines, source)
	lines = appendRoleLines(lines, source)

	if sourceNeedsUnverifiedLine(source) {
		lines = append(lines, identityUnverifiedLine)
	}

	return lines
}

func appendAccountLines(lines []string, source Source) []string {
	if source.AccountID != "" {
		lines = append(lines, "account: "+source.AccountID)
	}

	if source.SSOAccountID != "" {
		if source.AccountID != "" {
			return append(lines, "sso account: "+source.SSOAccountID)
		}

		return append(lines, "account: "+source.SSOAccountID)
	}

	return lines
}

func appendRoleLines(lines []string, source Source) []string {
	if source.RoleARN != "" {
		lines = append(lines, "role: "+source.RoleARN)
	}

	if source.SSORoleName != "" {
		if source.RoleARN != "" {
			return append(lines, "sso role: "+source.SSORoleName)
		}

		return append(lines, "role: "+source.SSORoleName)
	}

	return lines
}

func sourceNeedsUnverifiedLine(source Source) bool {
	return source.CredentialProcess ||
		source.CredentialSource != "" ||
		source.SourceProfile != ""
}
