package gate

import (
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"io"
	"strings"
	"unicode"
)

const (
	reasonPolicyDeny           = "the policy denies this subject"
	reasonProfileNotApproved   = "this profile is not approved on this machine"
	reasonGrantCannotAuthorize = "a grant cannot authorize this subject"
	allowGuidanceAllowFirst    = "allow it first, for a limited time, then retry"
	allowGuidanceHeadless      = "the grant covers later invocations on this machine, " +
		"including headless ones, until it expires"
	allowUsageLine                 = "usage: aws-guard allow <profile> [--ttl <duration>]"
	defaultAllowTTLWhenUnspecified = "5m"
)

func writeProfileRefusal(
	stderr io.Writer,
	subj subject.Subject,
	inv invocation.Invocation,
	decision policy.Decision,
	prof policy.Profile,
) {
	output.Cancel(stderr, cancelReasonForDecision(subj, decision))

	if subj.Kind == subject.NamedProfile && subj.Name != "" {
		output.Detail(stderr, "profile: "+subj.Name)
	}

	writeCommandDetail(stderr, inv)
	output.Detail(stderr, "policy: "+string(decision.Action))
	output.Detail(stderr, "reason: "+refusalReasonLine(subj, decision))

	if decision.Action == policy.ActionPrompt && subj.Kind == subject.NamedProfile {
		writeAllowGuidance(stderr, subj.Name, prof)
	}
}

func cancelReasonForDecision(subj subject.Subject, decision policy.Decision) string {
	return refusalReasonLine(subj, decision)
}

func refusalReasonLine(subj subject.Subject, decision policy.Decision) string {
	switch decision.Action {
	case policy.ActionAllow:
		return reasonPolicyDeny
	case policy.ActionDeny:
		return reasonPolicyDeny
	case policy.ActionPrompt:
		if subj.Kind == subject.NamedProfile {
			return reasonProfileNotApproved
		}

		return reasonGrantCannotAuthorize
	default:
		return reasonPolicyDeny
	}
}

func writeCommandDetail(stderr io.Writer, inv invocation.Invocation) {
	if !inv.OK || inv.Command == "" {
		return
	}

	line := "command: " + inv.Command
	if inv.Operation != "" {
		line += " " + inv.Operation
	}

	output.Detail(stderr, line)
}

func writeAllowGuidance(stderr io.Writer, profile string, prof policy.Profile) {
	output.Detail(stderr, allowGuidanceAllowFirst)
	output.Detail(stderr, allowGuidanceHeadless)
	output.Detail(stderr, allowCommandLine(profile, prof))
}

func allowCommandLine(profile string, prof policy.Profile) string {
	if !embeddableAllowProfile(profile) {
		return allowUsageLine
	}

	ttl := defaultAllowTTLWhenUnspecified
	if prof.HasApprovalTTL {
		ttl = prof.ApprovalTTL.String()
	}

	return "aws-guard allow " + profile + " --ttl " + ttl
}

func embeddableAllowProfile(name string) bool {
	if name == "" {
		return false
	}

	if strings.HasPrefix(name, "-") {
		return false
	}

	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}

	return true
}
