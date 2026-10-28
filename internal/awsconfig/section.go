package awsconfig

import (
	"maps"
	"strings"
)

const sectionNameParts = 2

func buildFromConfig(
	path string,
	data []byte,
) (map[string]Section, map[string]Section, map[string]Section, error) {
	doc, err := parseINI(path, data)
	if err != nil {
		return nil, nil, nil, err
	}

	doc = applyDefaults(doc)

	profiles := map[string]Section{}
	ssoSessions := map[string]Section{}
	services := map[string]Section{}

	for _, name := range doc.sectionOrd {
		sec, matErr := materializeSection(path, *doc.sections[name])
		if matErr != nil {
			return nil, nil, nil, matErr
		}

		routeSection(name, sec, profiles, ssoSessions, services)
	}

	return profiles, ssoSessions, services, nil
}

func buildFromCredentials(path string, data []byte) (map[string]Section, error) {
	doc, err := parseINI(path, data)
	if err != nil {
		return nil, err
	}

	doc = applyDefaults(doc)

	out := map[string]Section{}

	for _, name := range doc.sectionOrd {
		sec, matErr := materializeSection(path, *doc.sections[name])
		if matErr != nil {
			return nil, matErr
		}

		out[name] = sec
	}

	return out, nil
}

func routeSection(
	name string,
	sec Section,
	profiles, ssoSessions, services map[string]Section,
) {
	parts := splitSectionName(name)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "profile":
		if len(parts) != sectionNameParts {
			return
		}

		profiles[parts[1]] = sec
	case "sso-session":
		if len(parts) != sectionNameParts {
			return
		}

		ssoSessions[parts[1]] = sec
	case "services":
		if len(parts) != sectionNameParts {
			return
		}

		services[parts[1]] = sec
	default:
		if name == "default" {
			profiles["default"] = sec
		}
	}
}

func mergeProfiles(base, cred map[string]Section) map[string]Section {
	out := map[string]Section{}

	maps.Copy(out, base)

	for name, credSec := range cred {
		existing, ok := out[name]
		if !ok {
			out[name] = credSec

			continue
		}

		out[name] = mergeSections(existing, credSec)
	}

	return out
}

func mergeSections(configSec, credSec Section) Section {
	out := emptySection()

	maps.Copy(out.Settings, configSec.Settings)

	for block, nested := range configSec.Nested {
		out.Nested[block] = copyStringMap(nested)
	}

	out.HasAccessKeyID = configSec.HasAccessKeyID
	out.HasSecretAccessKey = configSec.HasSecretAccessKey
	out.HasSessionToken = configSec.HasSessionToken
	out.HasCredentialProcess = configSec.HasCredentialProcess

	maps.Copy(out.Settings, credSec.Settings)

	for block, nested := range credSec.Nested {
		out.Nested[block] = copyStringMap(nested)
	}

	if credSec.HasAccessKeyID {
		out.HasAccessKeyID = true
	}

	if credSec.HasSecretAccessKey {
		out.HasSecretAccessKey = true
	}

	if credSec.HasSessionToken {
		out.HasSessionToken = true
	}

	if credSec.HasCredentialProcess {
		out.HasCredentialProcess = true
	}

	return out
}

func copyStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	maps.Copy(out, in)

	return out
}

func splitSectionName(name string) []string {
	return shlexFields(name)
}

func shlexFields(input string) []string {
	var fields []string

	var cur strings.Builder

	inQuote := false
	quoteChar := rune(0)

	flush := func() {
		if cur.Len() > 0 {
			fields = append(fields, cur.String())
			cur.Reset()
		}
	}

	for _, runeVal := range input {
		switch {
		case inQuote:
			if runeVal == quoteChar {
				inQuote = false

				continue
			}

			cur.WriteRune(runeVal)
		case runeVal == '"' || runeVal == '\'':
			inQuote = true
			quoteChar = runeVal
		case runeVal == ' ' || runeVal == '\t':
			if cur.Len() > 0 {
				flush()
			}
		default:
			cur.WriteRune(runeVal)
		}
	}

	if inQuote {
		return nil
	}

	if cur.Len() > 0 {
		fields = append(fields, cur.String())
	}

	return fields
}
