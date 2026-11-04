package cli

import (
	"aws-cli-guard/internal/output"
	"io"
)

const (
	configCommand           = "config"
	configSubValidate       = "validate"
	configValidateArgsExact = 3
	reasonConfigUsage       = "usage: aws-guard config validate"
)

func runConfig(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != configValidateArgsExact || args[2] != configSubValidate {
		output.Error(stderr, reasonConfigUsage)

		return output.CancelStatus
	}

	_, status := loadPolicyForCLI(env, stderr)
	if status != 0 {
		return status
	}

	output.Detail(stdout, "policy: valid")

	return 0
}
