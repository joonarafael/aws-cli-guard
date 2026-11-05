package policy

import "aws-cli-guard/internal/subject"

// EvaluateForCaller returns the profile action for subj. When recognized is true and
// subj is a named profile, agent rules in [agents] may override the normal profile policy.
func EvaluateForCaller(pol Policy, subj subject.Subject, recognized bool) Decision {
	if recognized && subj.Kind == subject.NamedProfile {
		if dec, ok := evaluateRecognizedAgent(pol, subj.Name); ok {
			return dec
		}
	}

	return Evaluate(pol, subj)
}

func evaluateRecognizedAgent(pol Policy, name string) (Decision, bool) {
	if pol.AgentProfiles != nil {
		if action, ok := pol.AgentProfiles[name]; ok {
			return decisionFromAgentProfile(action), true
		}
	}

	if pol.HasAgentDefault {
		return decisionFromAgentDefault(pol.AgentDefault), true
	}

	return Decision{Action: ActionDeny, Basis: Basis(0)}, false
}

func decisionFromAgentProfile(action Action) Decision {
	if knownAction(action) {
		return Decision{Action: action, Basis: BasisAgentProfile}
	}

	return Decision{Action: ActionDeny, Basis: BasisAgentProfile}
}

func decisionFromAgentDefault(action Action) Decision {
	if knownAction(action) {
		return Decision{Action: action, Basis: BasisAgentDefault}
	}

	return Decision{Action: ActionDeny, Basis: BasisAgentDefault}
}

// IsAgentBasis reports whether basis selects an [agents] rule.
func IsAgentBasis(b Basis) bool {
	switch b {
	case BasisAgentProfile, BasisAgentDefault:
		return true
	case BasisProfile, BasisDefault, BasisCredentialExport, BasisEnvironment:
		return false
	default:
		return false
	}
}
