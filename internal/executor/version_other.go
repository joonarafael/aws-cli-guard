//go:build !unix

package executor

import "io"

// PolicyGate decides whether gate.Decide runs before exec.
// Non-Unix builds skip the version probe until process replacement exists.
func PolicyGate(_ string, _ []string, _ io.Writer) (bool, bool) {
	return true, true
}
