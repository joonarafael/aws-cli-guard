package policy

// ExportDecision is the [credential_export] outcome for one invocation.
type ExportDecision struct {
	Applies bool
	Action  Action
	Basis   Basis
}

// EvaluateExport applies [credential_export] when printsAccessKeys is true.
func EvaluateExport(pol Policy, printsAccessKeys bool) ExportDecision {
	if !printsAccessKeys {
		return ExportDecision{Applies: false, Action: "", Basis: Basis(0)}
	}

	action := pol.CredentialExport
	if !knownAction(action) {
		action = ActionDeny
	}

	return ExportDecision{
		Applies: true,
		Action:  action,
		Basis:   BasisCredentialExport,
	}
}

// Refuses reports whether the export gate blocks execution.
func (d ExportDecision) Refuses() bool {
	return d.Applies && d.Action != ActionAllow
}
