//go:build !unix

package cli

import (
	"aws-cli-guard/internal/output"
	"io"
)

const reasonPersistentDisablePlatform = "persistent disable is not supported on this platform"

func runPersistentDisableScript(_ string, _, stderr io.Writer) int {
	output.Error(stderr, reasonPersistentDisablePlatform)

	return output.CancelStatus
}
