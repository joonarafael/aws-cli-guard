//go:build unix

package files

import (
	"strconv"
	"time"
)

const maxPartialAttempts = 100

func partialPathCandidates(livePath string, now time.Time) []string {
	stamp := now.UTC().Format("20060102T150405.000000000Z")
	base := livePath + ".partial." + stamp

	out := make([]string, 0, maxPartialAttempts)
	out = append(out, base)

	for i := 2; i <= maxPartialAttempts; i++ {
		out = append(out, base+"."+strconv.Itoa(i))
	}

	return out
}
