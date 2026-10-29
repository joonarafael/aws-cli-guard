//go:build unix

package files

import (
	"strconv"
	"time"
)

const maxSiblingAttempts = 100

func siblingPathCandidates(livePath, marker string, now time.Time) []string {
	stamp := now.UTC().Format("20060102T150405.000000000Z")
	base := livePath + "." + marker + "." + stamp

	out := make([]string, 0, maxSiblingAttempts)
	out = append(out, base)

	for i := 2; i <= maxSiblingAttempts; i++ {
		out = append(out, base+"."+strconv.Itoa(i))
	}

	return out
}

func partialPathCandidates(livePath string, now time.Time) []string {
	return siblingPathCandidates(livePath, "partial", now)
}

func corruptPathCandidates(livePath string, now time.Time) []string {
	return siblingPathCandidates(livePath, "corrupt", now)
}
