package policy

import (
	"time"
	"unicode"
)

func validatePolicy(path string, doc document) (Policy, error) {
	out := Policy{
		Path:             path,
		Version:          doc.Version,
		DefaultAction:    ActionPrompt,
		CredentialExport: ActionPrompt,
		Profiles:         map[string]Profile{},
	}

	defaultAction, err := resolveDefaults(doc.Defaults)
	if err != nil {
		return Policy{}, err
	}

	out.DefaultAction = defaultAction

	exportAction, err := resolveCredentialExport(doc.CredentialExport)
	if err != nil {
		return Policy{}, err
	}

	out.CredentialExport = exportAction

	for name, prof := range doc.Profiles {
		keyPrefix := "profiles." + name

		nameErr := validateProfileName(name, keyPrefix)
		if nameErr != nil {
			return Policy{}, nameErr
		}

		profileOut, profErr := validateProfile(prof, keyPrefix)
		if profErr != nil {
			return Policy{}, profErr
		}

		out.Profiles[name] = profileOut
	}

	return out, nil
}

func resolveDefaults(sec *defaultsDocument) (Action, error) {
	if sec == nil {
		return ActionPrompt, nil
	}

	if sec.Action == nil {
		return "", newError(KindInvalid, "", "defaults.action", 0, 0, nil)
	}

	return parseAction(*sec.Action, "defaults.action")
}

func resolveCredentialExport(sec *credentialExportDocument) (Action, error) {
	if sec == nil {
		return ActionPrompt, nil
	}

	if sec.Action == nil {
		return "", newError(KindInvalid, "", "credential_export.action", 0, 0, nil)
	}

	return parseAction(*sec.Action, "credential_export.action")
}

func validateProfile(prof profileDocument, keyPrefix string) (Profile, error) {
	if prof.Action == nil {
		return Profile{}, newError(KindInvalid, "", keyPrefix+".action", 0, 0, nil)
	}

	action, err := parseAction(*prof.Action, keyPrefix+".action")
	if err != nil {
		return Profile{}, err
	}

	out := Profile{
		Action:         action,
		ApprovalTTL:    0,
		HasApprovalTTL: false,
	}

	if prof.ApprovalTTL == nil {
		return out, nil
	}

	ttl, ttlErr := parseApprovalTTL(*prof.ApprovalTTL, keyPrefix+".approval_ttl")
	if ttlErr != nil {
		return Profile{}, ttlErr
	}

	out.ApprovalTTL = ttl
	out.HasApprovalTTL = true

	return out, nil
}

func parseAction(raw, key string) (Action, error) {
	switch Action(raw) {
	case ActionAllow, ActionPrompt, ActionDeny:
		return Action(raw), nil
	default:
		return "", newError(KindInvalid, "", key, 0, 0, nil)
	}
}

func parseApprovalTTL(raw, key string) (time.Duration, error) {
	dur, err := time.ParseDuration(raw)
	if err != nil || dur <= 0 {
		return 0, newError(KindInvalid, "", key, 0, 0, nil)
	}

	return dur, nil
}

func validateProfileName(name, keyPrefix string) error {
	if name == "" {
		return newError(KindInvalid, "", keyPrefix, 0, 0, nil)
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return newError(KindInvalid, "", keyPrefix, 0, 0, nil)
		}
	}

	return nil
}
