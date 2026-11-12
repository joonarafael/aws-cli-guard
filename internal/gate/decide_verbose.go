package gate

import (
	"aws-cli-guard/internal/agent"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/subject"
	"aws-cli-guard/internal/verbose"
	"io"
	"time"
)

func homeLabel(env []string) string {
	label, ok := paths.HomeOrigin(env)
	if !ok {
		return ""
	}

	return label
}

func subjectLabel(subj subject.Subject) string {
	switch subj.Kind {
	case subject.NamedProfile:
		if subj.Name != "" {
			return "named profile " + subj.Name
		}

		return "named profile"
	case subject.Environment:
		return "environment"
	case subject.Unresolved:
		if subj.Reason != "" {
			return "unresolved " + subj.Reason
		}

		return "unresolved"
	default:
		return ""
	}
}

func commandLabel(inv invocation.Invocation) string {
	if !inv.OK || inv.Command == "" {
		return ""
	}

	if inv.Operation != "" {
		return inv.Command + " " + inv.Operation
	}

	return inv.Command
}

func callerLabel(ctx agent.Context) string {
	if ctx.Recognized() {
		return "recognized"
	}

	return "not recognized"
}

func ruleLabel(basis policy.Basis, action policy.Action) string {
	label := basis.Label()
	if label == "" {
		return string(action)
	}

	return label + " " + string(action)
}

func writePreRuleRefuse(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	realAWS string,
) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // partial decision block.
		Home:    homeLabel(env),
		Subject: subjectLabel(subj),
		Command: commandLabel(inv),
		AWS:     realAWS,
		Outcome: "refuse",
	})
}

func writeLoginHandoff(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	realAWS string,
) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // partial decision block.
		Home:    homeLabel(env),
		Command: commandLabel(inv),
		AWS:     realAWS,
		Outcome: "login handoff",
	})
}

type profileGrantResult struct {
	permits   bool
	grantWord string
}

func evaluateProfileGrant(
	subj subject.Subject,
	decision policy.Decision,
	grantList []grants.Grant,
	env []string,
) profileGrantResult {
	var fingerprint grants.Source

	var fingerprintErr error

	if decision.Action == policy.ActionPrompt && subj.Kind == subject.NamedProfile &&
		subj.Name != "" {
		fingerprint, fingerprintErr = grants.Fingerprint(env, subj.Name)
	}

	grantDecision := policy.EvaluateGrant(
		decision,
		subj,
		grantList,
		time.Now().UTC(),
		fingerprint,
		fingerprintErr,
	)

	word := ""
	if decision.Action == policy.ActionPrompt && subj.Kind == subject.NamedProfile &&
		subj.Name != "" {
		word = grantWordForProfile(
			subj.Name,
			grantList,
			time.Now().UTC(),
			fingerprint,
			fingerprintErr,
		)
	}

	return profileGrantResult{permits: grantDecision.Permits(), grantWord: word}
}

func grantWordForProfile(
	profile string,
	list []grants.Grant,
	now time.Time,
	live grants.Source,
	readErr error,
) string {
	var matched []grants.Grant

	for _, grant := range list {
		if grant.Profile == profile {
			matched = append(matched, grant)
		}
	}

	if len(matched) == 0 {
		return "absent"
	}

	now = now.UTC()

	for _, grant := range matched {
		if grants.GrantStatus(grant, now, live, readErr) == grants.GrantStatusCurrent {
			return "current"
		}
	}

	return grants.GrantStatus(matched[0], now, live, readErr).Label()
}

func writeDecisionEvent(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	caller agent.Context,
	realAWS string,
	rule string,
	grantWord string,
	outcome string,
) {
	verbose.Write(stderr, level, verbose.Event{
		Home:    homeLabel(env),
		Subject: subjectLabel(subj),
		Command: commandLabel(inv),
		Rule:    rule,
		Grant:   grantWord,
		Stdin:   caller.StdinLabel(),
		Caller:  callerLabel(caller),
		AWS:     realAWS,
		Outcome: outcome,
	})
}

func writeSessionOffEvent(
	stderr io.Writer,
	level settings.Level,
	env []string,
	inv invocation.Invocation,
	subj subject.Subject,
	caller agent.Context,
	realAWS string,
) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // session off block.
		Home:    homeLabel(env),
		Subject: subjectLabel(subj),
		Command: commandLabel(inv),
		Stdin:   caller.StdinLabel(),
		Caller:  callerLabel(caller),
		AWS:     realAWS,
		Outcome: "session off",
	})
}

func writeUnclassifiedRefuse(
	stderr io.Writer,
	level settings.Level,
	env []string,
	realAWS string,
) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // partial decision block.
		Home:    homeLabel(env),
		AWS:     realAWS,
		Outcome: "refuse",
	})
}
