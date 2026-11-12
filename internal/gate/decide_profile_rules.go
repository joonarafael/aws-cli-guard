package gate

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/subject"
	"io"
)

func applyProfileRules(
	inv invocation.Invocation,
	subj subject.Subject,
	pol policy.Policy,
	grantList []grants.Grant,
	env []string,
	realAWS string,
	level settings.Level,
	caller agent.Context,
	stderr io.Writer,
) bool {
	decision := policy.EvaluateForCaller(pol, subj, caller.Recognized())

	if policy.IsAgentBasis(decision.Basis) && decision.Action == policy.ActionDeny {
		return refuseAgentProfile(stderr, env, inv, subj, pol, level, caller, realAWS, decision)
	}

	export := policy.EvaluateExport(pol, inv.PrintsAccessKeys)
	if export.Refuses() {
		output.Cancel(stderr, policy.ReasonExportRefused)
		logExportRefuse(stderr, level, env, inv, subj, caller, realAWS, export)

		return false
	}

	grantResult := evaluateProfileGrant(subj, decision, grantList, env)
	if grantResult.permits {
		logDecision(
			stderr,
			level,
			env,
			inv,
			subj,
			caller,
			realAWS,
			decision,
			grantResult.grantWord,
			"allow",
		)

		return true
	}

	prof := pol.Profiles[subj.Name]
	writeProfileRefusal(stderr, env, subj, inv, decision, prof)
	logDecision(
		stderr,
		level,
		env,
		inv,
		subj,
		caller,
		realAWS,
		decision,
		grantResult.grantWord,
		"refuse",
	)

	return false
}

func refuseAgentProfile(
	stderr io.Writer,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	pol policy.Policy,
	level settings.Level,
	caller agent.Context,
	realAWS string,
	decision policy.Decision,
) bool {
	prof := pol.Profiles[subj.Name]
	writeProfileRefusal(stderr, env, subj, inv, decision, prof)
	logDecision(stderr, level, env, inv, subj, caller, realAWS, decision, "", "refuse")

	return false
}

func logDecision(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	caller agent.Context,
	realAWS string,
	decision policy.Decision,
	grantWord string,
	outcome string,
) {
	writeDecisionEvent(
		stderr,
		level,
		env,
		inv,
		subj,
		caller,
		realAWS,
		ruleLabel(decision.Basis, decision.Action),
		grantWord,
		outcome,
	)
}

func logExportRefuse(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	caller agent.Context,
	realAWS string,
	export policy.ExportDecision,
) {
	writeDecisionEvent(
		stderr,
		level,
		env,
		inv,
		subj,
		caller,
		realAWS,
		ruleLabel(export.Basis, export.Action),
		"",
		"refuse",
	)
}
