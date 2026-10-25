package subject

import (
	"aws-cli-guard/internal/environ"
	"strings"
)

const (
	reasonConflictingProfiles     = "conflicting-profiles"
	reasonProfileFlagMissingValue = "profile-flag-missing-value"
	reasonEmptyProfileName        = "empty-profile-name"
	//nolint:gosec // G101: policy reason label, not a credential value
	reasonPartialEnvCredentials = "partial-env-credentials"
	//nolint:gosec // G101: policy reason label, not a credential value
	reasonAmbientCredentials = "ambient-credentials"
)

// Resolve determines the credential subject from process arguments and environment.
func Resolve(args []string, env []string) Subject {
	flags := parseProfileFlags(args)
	if !flags.ok {
		return Subject{Kind: Unresolved, Name: "", Reason: flags.reason}
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

	name := flags.name
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

type profileFlagParse struct {
	ok     bool
	name   string
	reason string
}

func parseProfileFlags(args []string) profileFlagParse {
	var names []string

	for argIndex := 1; argIndex < len(args); argIndex++ {
		arg := args[argIndex]
		if strings.HasPrefix(arg, "--profile=") {
			profileName := arg[len("--profile="):]
			if profileName == "" {
				return profileFlagParse{ok: false, name: "", reason: reasonEmptyProfileName}
			}

			names = append(names, profileName)

			continue
		}

		if arg != "--profile" {
			continue
		}

		if argIndex+1 >= len(args) {
			return profileFlagParse{ok: false, name: "", reason: reasonProfileFlagMissingValue}
		}

		argIndex++

		profileName := args[argIndex]
		if profileName == "" {
			return profileFlagParse{ok: false, name: "", reason: reasonEmptyProfileName}
		}

		if strings.HasPrefix(profileName, "-") {
			return profileFlagParse{ok: false, name: "", reason: reasonProfileFlagMissingValue}
		}

		names = append(names, profileName)
	}

	if len(names) == 0 {
		return profileFlagParse{ok: true, name: "", reason: ""}
	}

	first := names[0]
	for _, other := range names[1:] {
		if other != first {
			return profileFlagParse{ok: false, name: "", reason: reasonConflictingProfiles}
		}
	}

	return profileFlagParse{ok: true, name: first, reason: ""}
}
