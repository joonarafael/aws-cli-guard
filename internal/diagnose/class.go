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
