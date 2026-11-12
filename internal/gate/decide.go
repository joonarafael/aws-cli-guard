package gate

import (
	"aws-cli-guard/internal/invocation"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/verbose"
	"io"
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
	reasonSessionsUnusable  = "the session file is unusable"
	reasonSessionsPathUnset = "the session path could not be resolved"
	reasonSessionsReadFail  = "the session file could not be read"
)

// Decide reports whether the wrapper may exec the real AWS CLI for args and env.
func Decide(args []string, env []string, realAWS string, stderr io.Writer) bool {
	envLevel, err := verbose.EnvLevel(env)
	if err != nil {
		output.Cancel(stderr, verbose.ReasonInvalid)

		return false
	}

	inv := invocation.Parse(args)
	if inv.IsLogin() {
		writeLoginHandoff(stderr, envLevel, env, inv, realAWS)

		return true
	}

	loaded, ok := decideLoadFiles(args, env, realAWS, stderr, envLevel, inv)
	if !ok {
		return false
	}

	return decideAfterLoads(
		inv,
		loaded.subj,
		loaded.pol,
		loaded.grantList,
		env,
		realAWS,
		loaded.level,
		stderr,
	)
}
