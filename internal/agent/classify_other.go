//go:build !unix

package agent

import "os"

func classifyStdin(_ *os.File) Stdin {
	return StdinUndetermined
}
