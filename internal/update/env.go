package update

import (
	"slices"
	"strings"
)

// StripAWSCredentials removes AWS credential env vars before HTTP requests.
func StripAWSCredentials(env []string) []string {
	credEnvNames := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"AWS_SECURITY_TOKEN",
		"AWS_PROFILE",
		"AWS_DEFAULT_PROFILE",
	}
	out := make([]string, 0, len(env))

	for _, entry := range env {
		name, _, ok := strings.Cut(entry, "=")
		if !ok {
			out = append(out, entry)

			continue
		}

		if containsName(credEnvNames, name) {
			continue
		}

		out = append(out, entry)
	}

	return out
}

func containsName(names []string, want string) bool {
	return slices.Contains(names, want)
}
