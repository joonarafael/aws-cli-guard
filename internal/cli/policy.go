package cli

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"io"
	"sort"
	"strconv"
)

const (
	policyCommand          = "policy"
	policySubList          = "list"
	policySubExplain       = "explain"
	policyMinArgsWithSub   = 3
	policyListArgsExact    = 3
	policyExplainArgsExact = 4
	reasonPolicyUsage      = "usage: aws-guard policy list|explain <profile>"
)

func runPolicy(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) < policyMinArgsWithSub {
		output.Error(stderr, reasonPolicyUsage)

		return output.CancelStatus
	}

	switch args[2] {
	case policySubList:
		return runPolicyList(args, env, stdout, stderr)
	case policySubExplain:
		return runPolicyExplain(args, env, stdout, stderr)
	default:
		output.Error(stderr, reasonPolicyUsage)

		return output.CancelStatus
	}
}

func runPolicyList(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != policyListArgsExact {
		output.Error(stderr, reasonPolicyUsage)

		return output.CancelStatus
	}

	pol, status := loadPolicyForCLI(env, stderr)
	if status != 0 {
		return status
	}

	writePolicyList(stdout, pol)

	return 0
}

func writePolicyList(stdout io.Writer, pol policy.Policy) {
	output.Detail(stdout, "version: "+strconv.Itoa(pol.Version))
	output.Detail(stdout, "defaults: "+string(pol.DefaultAction))

	envDecision := policy.Evaluate(pol, subject.Subject{
		Kind: subject.Environment, Name: "", Reason: "",
	})
	output.Detail(stdout, "environment: "+string(envDecision.Action))
	output.Detail(stdout, "environment_basis: "+envDecision.Basis.Label())
	output.Detail(stdout, "credential_export: "+string(pol.CredentialExport))

	if pol.HasEnvironmentMarkers {
		output.Detail(stdout, "environment_markers: "+strconv.FormatBool(pol.EnvironmentMarkers))
	}

	names := make([]string, 0, len(pol.Profiles))
	for name := range pol.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		prof := pol.Profiles[name]
		output.Detail(stdout, "profile: "+name)
		output.Detail(stdout, "action: "+string(prof.Action))

		if prof.HasApprovalTTL {
			output.Detail(stdout, "approval_ttl: "+prof.ApprovalTTL.String())
		}
	}
}

func runPolicyExplain(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != policyExplainArgsExact {
		output.Error(stderr, reasonPolicyUsage)

		return output.CancelStatus
	}

	profile := args[3]
	if profile == "" || profile[0] == '-' {
		output.Error(stderr, reasonPolicyUsage)

		return output.CancelStatus
	}

	if !validProfileName(profile) {
		output.Error(stderr, reasonProfileInvalid)

		return output.CancelStatus
	}

	pol, status := loadPolicyForCLI(env, stderr)
	if status != 0 {
		return status
	}

	writePolicyExplain(stdout, env, pol, profile)

	return 0
}

func writePolicyExplain(stdout io.Writer, env []string, pol policy.Policy, profile string) {
	output.Detail(stdout, "profile: "+profile)

	source, err := grants.Fingerprint(env, profile)
	if err == nil {
		for _, line := range grants.IdentityLines(source) {
			output.Detail(stdout, line)
		}
	}

	subj := subject.Subject{Kind: subject.NamedProfile, Name: profile, Reason: ""}
	decision := policy.Evaluate(pol, subj)

	output.Detail(stdout, "action: "+string(decision.Action))
	output.Detail(stdout, "basis: "+decision.Basis.Label())

	prof, hasProf := pol.Profiles[profile]
	if hasProf && prof.HasApprovalTTL {
		output.Detail(stdout, "approval_ttl: "+prof.ApprovalTTL.String())
	}

	writeGrantRequirement(stdout, decision.Action)
	output.Detail(stdout, "credential_export: "+string(pol.CredentialExport))
}

func writeGrantRequirement(stdout io.Writer, action policy.Action) {
	if action == policy.ActionPrompt {
		output.Detail(stdout, "grant: required")

		return
	}

	output.Detail(stdout, "grant: not used")
}
