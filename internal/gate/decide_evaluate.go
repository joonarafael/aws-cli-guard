package gate

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/subject"
	"io"
)

func decideAfterLoads(
	inv invocation.Invocation,
	subj subject.Subject,
	pol policy.Policy,
	grantList []grants.Grant,
	env []string,
	realAWS string,
	level settings.Level,
	stderr io.Writer,
) bool {
	caller, ok := observeCaller(env, pol, stderr)
	if !ok {
		writeUnclassifiedRefuse(stderr, level, env, realAWS)

		return false
	}

	if permit, stopped := trySessionOffHandoff(
		inv,
		subj,
		env,
		realAWS,
		level,
		caller,
		stderr,
	); stopped {
		return permit
	}

	return applyProfileRules(inv, subj, pol, grantList, env, realAWS, level, caller, stderr)
}

func trySessionOffHandoff(
	inv invocation.Invocation,
	subj subject.Subject,
	env []string,
	realAWS string,
	level settings.Level,
	caller agent.Context,
	stderr io.Writer,
) (bool, bool) {
	if caller.Recognized() {
		return false, false
	}

	skipped, sessOK := sessionOffPermits(env, stderr)
	if !sessOK {
		writePreRuleRefuse(stderr, level, env, inv, subj, realAWS)

		return false, true
	}

	if !skipped {
		return false, false
	}

	output.Warning(stderr, session.SessionOffWarning)
	writeSessionOffEvent(stderr, level, env, inv, subj, caller, realAWS)

	return true, true
}
