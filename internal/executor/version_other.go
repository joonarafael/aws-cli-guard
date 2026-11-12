//go:build !unix

package executor

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/verbose"
	"io"
)

// PolicyGate decides whether gate.Decide runs before exec.
// Non-Unix builds skip the version probe until process replacement exists.
func PolicyGate(_ string, env []string, stderr io.Writer) (bool, bool) {
	_, err := verbose.EnvLevel(env)
	if err != nil {
		output.Cancel(stderr, verbose.ReasonInvalid)

		return false, false
	}

	return true, true
}
