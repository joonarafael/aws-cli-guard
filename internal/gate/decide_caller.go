package gate

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/output"
	"io"
)

const reasonCallerUnclassified = "the caller context could not be classified"

func callerAllowsProfile(ctx agent.Context, stderr io.Writer) bool {
	if ctx.FollowsProfilePolicy() {
		return true
	}

	output.Cancel(stderr, reasonCallerUnclassified)

	return false
}
