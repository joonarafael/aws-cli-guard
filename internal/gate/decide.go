package gate

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/subject"
	"errors"
	"io"
	"time"
)

const (
	reasonPolicyUnusable    = "the policy file is unusable"
	reasonPolicyReadFail    = "the policy file could not be read"
	reasonGrantsUnusable    = "the grants file is unusable"
	reasonGrantsPathUnset   = "the grants path could not be resolved"
	reasonGrantsReadFail    = "the grants file could not be read"
	reasonSettingsUnusable  = "the settings file is unusable"
	reasonSettingsPathUnset = "the settings path could not be resolved"
	reasonSettingsReadFail  = "the settings file could not be read"
)

// Decide reports whether the wrapper may exec the real AWS CLI for args and env.
func Decide(args []string, env []string, stderr io.Writer) bool {
	inv := invocation.Parse(args)
	if inv.IsLogin() {
		return true
	}

	err := policy.Present(env)
	if err != nil {
		output.Cancel(stderr, policy.PresentReason(err))

		return false
	}

	subj := subject.Resolve(args, env)
	subject.WarnDefaultProfile(stderr, subj)

	setFile, err := settings.Load(env)
	if err != nil {
		return cancelSettingsLoad(stderr, err)
	}

	if setFile.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassSettings(), setFile.Path, setFile.Mode)
	}

	pol, err := policy.Load(env)
	if err != nil {
		return cancelPolicyLoad(stderr, err)
	}

	diagnose.WriteLooseMode(stderr, diagnose.ClassPolicy(), pol.Path, pol.Mode)

	grantFile, err := grants.Load(env)
	if err != nil {
		return cancelGrantsLoad(stderr, err)
	}

	if grantFile.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassGrants(), grantFile.Path, grantFile.Mode)
	}

	export := policy.EvaluateExport(pol, inv.PrintsAccessKeys)
	if export.Refuses() {
		output.Cancel(stderr, policy.ReasonExportRefused)

		return false
	}

	if permitsProfile(pol, subj, grantFile.Grants, env) {
		return true
	}

	decision := policy.Evaluate(pol, subj)
	prof := pol.Profiles[subj.Name]
	writeProfileRefusal(stderr, subj, inv, decision, prof)

	return false
}

func permitsProfile(
	pol policy.Policy,
	subj subject.Subject,
	grantList []grants.Grant,
	env []string,
) bool {
	decision := policy.Evaluate(pol, subj)

	var fingerprint grants.Source

	var fingerprintErr error

	if decision.Action == policy.ActionPrompt && subj.Kind == subject.NamedProfile &&
		subj.Name != "" {
		fingerprint, fingerprintErr = grants.Fingerprint(env, subj.Name)
	}

	grantDecision := policy.EvaluateGrant(
		decision,
		subj,
		grantList,
		time.Now().UTC(),
		fingerprint,
		fingerprintErr,
	)

	return grantDecision.Permits()
}

func cancelSettingsLoad(stderr io.Writer, err error) bool {
	if note, ok := settings.Corruption(err); ok {
		output.Cancel(stderr, reasonSettingsUnusable)
		diagnose.Write(stderr, note)

		return false
	}

	var setErr *settings.Error

	if errors.As(err, &setErr) && setErr.Kind == settings.KindPath {
		output.Cancel(stderr, reasonSettingsPathUnset)

		return false
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Cancel(stderr, reasonSettingsPathUnset)

		return false
	}

	output.Cancel(stderr, reasonSettingsReadFail)

	return false
}

func cancelPolicyLoad(stderr io.Writer, err error) bool {
	if note, ok := policy.Corruption(err); ok {
		output.Cancel(stderr, reasonPolicyUnusable)
		diagnose.Write(stderr, note)

		return false
	}

	if errors.Is(err, policy.ErrMissing) {
		output.Cancel(stderr, policy.ReasonMissing)

		return false
	}

	output.Cancel(stderr, reasonPolicyReadFail)

	return false
}

func cancelGrantsLoad(stderr io.Writer, err error) bool {
	if note, ok := grants.Corruption(err); ok {
		output.Cancel(stderr, reasonGrantsUnusable)
		diagnose.Write(stderr, note)

		return false
	}

	var grErr *grants.Error

	if errors.As(err, &grErr) && grErr.Kind == grants.KindPath {
		output.Cancel(stderr, reasonGrantsPathUnset)

		return false
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Cancel(stderr, reasonGrantsPathUnset)

		return false
	}

	output.Cancel(stderr, reasonGrantsReadFail)

	return false
}
