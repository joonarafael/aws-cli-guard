package subject

import (
	"aws-cli-guard/internal/environ"
	"aws-cli-guard/internal/invocation"
)

const (
	//nolint:gosec // G101: policy reason label, not a credential value
	reasonPartialEnvCredentials = "partial-env-credentials"
	//nolint:gosec // G101: policy reason label, not a credential value
	reasonAmbientCredentials = "ambient-credentials"
)

// Resolve determines the credential subject from process arguments and environment.
func Resolve(args []string, env []string) Subject {
	inv := invocation.Parse(args)
	if !inv.OK {
		return Subject{Kind: Unresolved, Name: "", Reason: inv.Reason}
	}

	accessKey, accessOK := environ.Lookup("AWS_ACCESS_KEY_ID", env)
	secretKey, secretOK := environ.Lookup("AWS_SECRET_ACCESS_KEY", env)

	if accessOK && secretOK && accessKey != "" && secretKey != "" {
		return Subject{Kind: Environment, Name: "", Reason: ""}
	}

	if accessOK && accessKey != "" && (!secretOK || secretKey == "") {
		return Subject{Kind: Unresolved, Name: "", Reason: reasonPartialEnvCredentials}
	}

	if ambientCredentials(env) {
		return Subject{Kind: Unresolved, Name: "", Reason: reasonAmbientCredentials}
	}

	name := ""
	if inv.HasProfile {
		name = inv.Profile
	}

	if name == "" {
		if profile, ok := environ.Lookup("AWS_PROFILE", env); ok && profile != "" {
			name = profile
		} else {
			name = "default"
		}
	}

	return Subject{Kind: NamedProfile, Name: name, Reason: ""}
}

func ambientCredentials(env []string) bool {
	keys := []string{
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI",
		"AWS_CONTAINER_CREDENTIALS_FULL_URI",
		"AWS_WEB_IDENTITY_TOKEN_FILE",
	}
	for _, key := range keys {
		if val, ok := environ.Lookup(key, env); ok && val != "" {
			return true
		}
	}

	return false
}
