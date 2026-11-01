package policy

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/subject"
	"time"
)

// GrantDecision is the profile action plus whether a current grant authorizes prompt.
type GrantDecision struct {
	Action  Action
	Current bool
}

// EvaluateGrant applies grant checking to decision. list is the loaded grants file;
// current is the credential-source fingerprint for a named profile; readErr is from
// Fingerprint when the AWS config could not be read.
func EvaluateGrant(
	decision Decision,
	subj subject.Subject,
	list []grants.Grant,
	now time.Time,
	current grants.Source,
	readErr error,
) GrantDecision {
	out := GrantDecision{Action: decision.Action, Current: false}

	if decision.Action != ActionPrompt {
		return out
	}

	if subj.Kind != subject.NamedProfile || subj.Name == "" || readErr != nil {
		return out
	}

	if grants.Current(list, subj.Name, now, current) {
		out.Current = true
	}

	return out
}

// Permits reports whether execution may proceed after grant checking.
func (d GrantDecision) Permits() bool {
	switch d.Action {
	case ActionAllow:
		return true
	case ActionPrompt:
		return d.Current
	case ActionDeny:
		return false
	default:
		return false
	}
}
