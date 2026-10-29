package invocation

import "strings"

const (
	commandConfigure = "configure"
	//nolint:gosec // G101: AWS CLI operation name, not a credential value
	operationExportCreds = "export-credentials"
	operationGet         = "get"

	settingAccessKeyID = "aws_access_key_id"
	//nolint:gosec // G101: AWS config key name, not a credential value
	settingSecretAccessKey = "aws_secret_access_key"

	settingSessionToken = "aws_session_token"
)

func printsAccessKeys(command, operation string, rest []string) bool {
	if command != commandConfigure {
		return false
	}

	switch operation {
	case operationExportCreds:
		return true
	case operationGet:
		return configureGetPrintsAccessKeys(rest)
	default:
		return false
	}
}

func configureGetPrintsAccessKeys(rest []string) bool {
	switch len(rest) {
	case 0:
		return true
	case 1:
		return secretSettingName(rest[0])
	default:
		return true
	}
}

func secretSettingName(token string) bool {
	_, setting, found := strings.Cut(token, ".")
	if !found {
		setting = token
	}

	return setting == settingAccessKeyID ||
		setting == settingSecretAccessKey ||
		setting == settingSessionToken
}
