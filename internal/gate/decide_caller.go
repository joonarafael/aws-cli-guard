package gate

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"io"
	"os"
)

const reasonCallerUnclassified = "the caller context could not be classified"

func observeCaller(env []string, pol policy.Policy, stderr io.Writer) (agent.Context, bool) {
	caller := agent.Observe(os.Stdin)

	caller = agent.ApplyMarkers(caller, env, pol.EnvironmentMarkers)

	if !callerAllowsProfile(caller, stderr) {
		return agent.Context{}, false
	}

	return caller, true
}

func callerAllowsProfile(ctx agent.Context, stderr io.Writer) bool {
	if ctx.FollowsProfilePolicy() {
		return true
	}

	output.Cancel(stderr, reasonCallerUnclassified)

	return false
}
