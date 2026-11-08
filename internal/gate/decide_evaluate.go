package gate

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/subject"
	"io"
)

func decideAfterLoads(
	inv invocation.Invocation,
	subj subject.Subject,
	pol policy.Policy,
	grantList []grants.Grant,
	env []string,
	stderr io.Writer,
) bool {
	caller, ok := observeCaller(env, pol, stderr)
	if !ok {
		return false
	}

	if !caller.Recognized() {
		skipped, sessOK := sessionOffPermits(env, stderr)
		if !sessOK {
			return false
		}

		if skipped {
			output.Warning(stderr, session.SessionOffWarning)

			return true
		}
	}

	decision := policy.EvaluateForCaller(pol, subj, caller.Recognized())

	if policy.IsAgentBasis(decision.Basis) && decision.Action == policy.ActionDeny {
		prof := pol.Profiles[subj.Name]
		writeProfileRefusal(stderr, env, subj, inv, decision, prof)

		return false
	}

	export := policy.EvaluateExport(pol, inv.PrintsAccessKeys)
	if export.Refuses() {
		output.Cancel(stderr, policy.ReasonExportRefused)

		return false
	}

	if permitsProfileDecision(subj, decision, grantList, env) {
		return true
	}

	prof := pol.Profiles[subj.Name]
	writeProfileRefusal(stderr, env, subj, inv, decision, prof)

	return false
}
