package diagnose

// Class names a guard-owned file type for diagnosis and recovery.
type Class struct {
	Name    string
	Recover string
}

// ClassPolicy returns the user policy file at config.toml.
func ClassPolicy() Class {
	return Class{
		Name:    "policy",
		Recover: "aws-guard recover policy",
	}
}

// ClassGrants returns the grants file at grants.json.
func ClassGrants() Class {
	return Class{
		Name:    "grants",
		Recover: "aws-guard recover grants",
	}
}

// ClassCache returns the disposable cache directory.
func ClassCache() Class {
	return Class{
		Name:    "cache",
		Recover: "aws-guard recover cache",
	}
}

// ClassSettings returns the preferences file at settings.toml.
func ClassSettings() Class {
	return Class{
		Name:    "settings",
		Recover: "aws-guard recover settings",
	}
}

// ClassSessions returns the session-off record file at sessions.json.
func ClassSessions() Class {
	return Class{
		Name:    "sessions",
		Recover: "aws-guard recover sessions",
	}
}
