package policy

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

func encodePolicyV4(pol Policy) []byte {
	var buf bytes.Buffer

	_, _ = fmt.Fprintf(&buf, "version = %d\n\n", readVersionAgentActions)

	writeActionSection(&buf, "defaults", pol.DefaultAction)
	_, _ = fmt.Fprint(&buf, "\n")

	if pol.HasEnvironmentAction {
		writeActionSection(&buf, "environment", pol.EnvironmentAction)
		_, _ = fmt.Fprint(&buf, "\n")
	}

	writeAgentsSections(&buf, pol)

	writeActionSection(&buf, "credential_export", pol.CredentialExport)
	_, _ = fmt.Fprint(&buf, "\n")

	writeProfileSections(&buf, pol.Profiles)

	out := buf.Bytes()

	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}

	return out
}

func writeActionSection(buf *bytes.Buffer, table string, action Action) {
	_, _ = fmt.Fprintf(buf, "[%s]\n", table)
	_, _ = fmt.Fprintf(buf, "action = %s\n", quoteTOMLString(string(action)))
}

func writeAgentsSections(buf *bytes.Buffer, pol Policy) {
	hasAgents := pol.HasEnvironmentMarkers || pol.HasAgentDefault || len(pol.AgentProfiles) > 0
	if !hasAgents {
		return
	}

	if pol.HasAgentDefault {
		_, _ = fmt.Fprint(buf, "[agents]\n")
		_, _ = fmt.Fprintf(buf, "default = %s\n", quoteTOMLString(string(pol.AgentDefault)))
	}

	if pol.HasEnvironmentMarkers {
		_, _ = fmt.Fprint(buf, "\n[agents.detection]\n")

		markers := "false"
		if pol.EnvironmentMarkers {
			markers = "true"
		}

		_, _ = fmt.Fprintf(buf, "environment_markers = %s\n", markers)
	}

	names := make([]string, 0, len(pol.AgentProfiles))
	for name := range pol.AgentProfiles {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		_, _ = fmt.Fprintf(buf, "\n[agents.profiles.%s]\n", dottedTableKey(name))

		action := pol.AgentProfiles[name]
		_, _ = fmt.Fprintf(buf, "action = %s\n", quoteTOMLString(string(action)))
	}

	_, _ = fmt.Fprint(buf, "\n")
}

func writeProfileSections(buf *bytes.Buffer, profiles map[string]Profile) {
	if len(profiles) == 0 {
		return
	}

	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		prof := profiles[name]

		_, _ = fmt.Fprintf(buf, "[profiles.%s]\n", dottedTableKey(name))
		_, _ = fmt.Fprintf(buf, "action = %s\n", quoteTOMLString(string(prof.Action)))

		if prof.HasApprovalTTL {
			_, _ = fmt.Fprintf(
				buf,
				"approval_ttl = %s\n",
				quoteTOMLString(prof.ApprovalTTL.String()),
			)
		}

		_, _ = fmt.Fprint(buf, "\n")
	}
}

func dottedTableKey(name string) string {
	if bareTOMLKey(name) {
		return name
	}

	return strconv.Quote(name)
}

func bareTOMLKey(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !isBareTOMLKeyRune(r) {
			return false
		}
	}

	return true
}

func isBareTOMLKeyRune(r rune) bool {
	return r == '-' || r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func quoteTOMLString(value string) string {
	return strconv.Quote(value)
}
