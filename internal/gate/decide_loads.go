package gate

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/subject"
	"aws-cli-guard/internal/verbose"
	"io"
)

type decideLoaded struct {
	subj      subject.Subject
	level     settings.Level
	pol       policy.Policy
	grantList []grants.Grant
}

func decideLoadFiles(
	args []string,
	env []string,
	realAWS string,
	stderr io.Writer,
	envLevel settings.Level,
	inv invocation.Invocation,
) (decideLoaded, bool) {
	err := policy.Present(env)
	if err != nil {
		var emptySubj subject.Subject
		writePreRuleRefuse(stderr, envLevel, env, inv, emptySubj, realAWS)
		output.Cancel(stderr, policy.PresentReason(err))

		return decideLoaded{}, false //nolint:exhaustruct_v5 // failed before resolve.
	}

	subj := subject.Resolve(args, env)
	subject.WarnDefaultProfile(stderr, subj)

	level, ok := loadSettingsLevel(stderr, env, envLevel, inv, subj, realAWS)
	if !ok {
		return decideLoaded{subj: subj}, false //nolint:exhaustruct_v5 // settings load failed.
	}

	pol, err := policy.Load(env)
	if err != nil {
		writePreRuleRefuse(stderr, level, env, inv, subj, realAWS)
		cancelPolicyLoad(stderr, err)

		return decideLoaded{ //nolint:exhaustruct_v5 // policy load failed.
			subj:  subj,
			level: level,
		}, false
	}

	diagnose.WriteLooseMode(stderr, diagnose.ClassPolicy(), pol.Path, pol.Mode)

	grantFile, err := grants.Load(env)
	if err != nil {
		writePreRuleRefuse(stderr, level, env, inv, subj, realAWS)
		cancelGrantsLoad(stderr, err)

		return decideLoaded{ //nolint:exhaustruct_v5 // grants load failed.
			subj:  subj,
			level: level,
			pol:   pol,
		}, false
	}

	if grantFile.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassGrants(), grantFile.Path, grantFile.Mode)
	}

	return decideLoaded{
		subj:      subj,
		level:     level,
		pol:       pol,
		grantList: grantFile.Grants,
	}, true
}

func loadSettingsLevel(
	stderr io.Writer,
	env []string,
	envLevel settings.Level,
	inv invocation.Invocation,
	subj subject.Subject,
	realAWS string,
) (settings.Level, bool) {
	setFile, err := settings.Load(env)
	if err != nil {
		writePreRuleRefuse(stderr, envLevel, env, inv, subj, realAWS)
		cancelSettingsLoad(stderr, err)

		return settings.LevelTight, false
	}

	level, err := verbose.Effective(env, setFile.Level)
	if err != nil {
		output.Cancel(stderr, verbose.ReasonInvalid)

		return settings.LevelTight, false
	}

	if setFile.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassSettings(), setFile.Path, setFile.Mode)
	}

	return level, true
}
