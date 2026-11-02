package policy

import "aws-cli-guard/internal/subject"

// Basis identifies which policy rule produced a decision.
type Basis int

const (
	// BasisProfile is a profile-specific override in [profiles].
	BasisProfile Basis = iota + 1
	// BasisDefault is [defaults] action for an unspecified profile or non-profile subject.
	BasisDefault
	// BasisCredentialExport is [credential_export] action.
	BasisCredentialExport
	// BasisEnvironment is [environment] action for subject.Environment.
	BasisEnvironment
)

// Decision is the profile-policy outcome for one credential subject.
type Decision struct {
	Action Action
	Basis  Basis
}

// Evaluate returns the profile action for subj under pol. It does not check grants
// or [credential_export]; callers apply EvaluateExport and EvaluateGrant separately.
func Evaluate(pol Policy, subj subject.Subject) Decision {
	switch subj.Kind {
	case subject.NamedProfile:
		return evaluateNamedProfile(pol, subj.Name)
	case subject.Environment:
		return evaluateEnvironment(pol)
	case subject.Unresolved:
		return decisionFromDefault(pol.DefaultAction)
	default:
		return Decision{Action: ActionDeny, Basis: Basis(0)}
	}
}

func evaluateEnvironment(pol Policy) Decision {
	if pol.HasEnvironmentAction {
		return decisionFromEnvironment(pol.EnvironmentAction)
	}

	return decisionFromDefault(pol.DefaultAction)
}

func decisionFromEnvironment(action Action) Decision {
	if knownAction(action) {
		return Decision{Action: action, Basis: BasisEnvironment}
	}

	return Decision{Action: ActionDeny, Basis: BasisEnvironment}
}

func evaluateNamedProfile(pol Policy, name string) Decision {
	if pol.Profiles != nil {
		if prof, ok := pol.Profiles[name]; ok {
			return decisionFromProfile(prof.Action)
		}
	}

	return decisionFromDefault(pol.DefaultAction)
}

func decisionFromProfile(action Action) Decision {
	if knownAction(action) {
		return Decision{Action: action, Basis: BasisProfile}
	}

	return Decision{Action: ActionDeny, Basis: BasisProfile}
}

func decisionFromDefault(action Action) Decision {
	if knownAction(action) {
		return Decision{Action: action, Basis: BasisDefault}
	}

	return Decision{Action: ActionDeny, Basis: BasisDefault}
}

func knownAction(action Action) bool {
	switch action {
	case ActionAllow, ActionPrompt, ActionDeny:
		return true
	default:
		return false
	}
}
