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
)

// Decision is the profile-policy outcome for one credential subject.
type Decision struct {
	Action Action
	Basis  Basis
}

// Evaluate returns the profile action for subj under pol. It does not check grants
// or [credential_export]; callers apply EvaluateExport and grant checks in later steps.
func Evaluate(pol Policy, subj subject.Subject) Decision {
	switch subj.Kind {
	case subject.NamedProfile:
		return evaluateNamedProfile(pol, subj.Name)
	case subject.Environment, subject.Unresolved:
		return decisionFromDefault(pol.DefaultAction)
	default:
		return Decision{Action: ActionDeny, Basis: Basis(0)}
	}
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
