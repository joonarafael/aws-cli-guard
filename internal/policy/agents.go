package policy

import "sort"

func resolveAgents(fileVersion int, sec *agentsDocument, out *Policy) error {
	if sec == nil {
		return nil
	}

	if fileVersion < readVersionAgent {
		return newError(KindInvalid, "", "agents", 0, 0, nil)
	}

	hasDefault := sec.Default != nil
	hasProfiles := len(sec.Profiles) > 0

	if fileVersion < readVersionAgentActions {
		if hasDefault || hasProfiles {
			return agentActionKeyInvalidKey(sec)
		}
	}

	if sec.Detection != nil {
		if sec.Detection.EnvironmentMarkers == nil {
			return newError(KindInvalid, "", "agents.detection.environment_markers", 0, 0, nil)
		}

		out.HasEnvironmentMarkers = true
		out.EnvironmentMarkers = *sec.Detection.EnvironmentMarkers
	}

	if fileVersion >= readVersionAgentActions && (hasDefault || hasProfiles) {
		return resolveAgentActionsV4(sec, out)
	}

	return nil
}

func resolveAgentActionsV4(sec *agentsDocument, out *Policy) error {
	if !out.EnvironmentMarkers {
		return agentActionKeyInvalidKey(sec)
	}

	err := resolveAgentDefault(sec, out)
	if err != nil {
		return err
	}

	return resolveAgentProfiles(sec, out)
}

func agentActionKeyInvalidKey(sec *agentsDocument) error {
	if sec.Default != nil {
		return newError(KindInvalid, "", "agents.default", 0, 0, nil)
	}

	names := make([]string, 0, len(sec.Profiles))
	for name := range sec.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	if len(names) == 0 {
		return newError(KindInvalid, "", "agents", 0, 0, nil)
	}

	return newError(KindInvalid, "", "agents.profiles."+names[0], 0, 0, nil)
}

func resolveAgentDefault(sec *agentsDocument, out *Policy) error {
	if sec.Default == nil {
		return nil
	}

	action, err := parseAction(*sec.Default, "agents.default")
	if err != nil {
		return err
	}

	out.HasAgentDefault = true
	out.AgentDefault = action

	return nil
}

func resolveAgentProfiles(sec *agentsDocument, out *Policy) error {
	for name, prof := range sec.Profiles {
		keyPrefix := "agents.profiles." + name

		nameErr := validateProfileName(name, keyPrefix)
		if nameErr != nil {
			return nameErr
		}

		if prof.Action == nil {
			return newError(KindInvalid, "", keyPrefix+".action", 0, 0, nil)
		}

		action, err := parseAction(*prof.Action, keyPrefix+".action")
		if err != nil {
			return err
		}

		out.AgentProfiles[name] = action
	}

	return nil
}
