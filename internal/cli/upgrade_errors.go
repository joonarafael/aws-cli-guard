package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"errors"
	"io"
)

func reportInstallOrUpgradeFailure(stderr io.Writer, env []string, err error) {
	programsReplaced := errors.Is(err, install.ErrUpgradePartial)

	if writePreflightDiagnosis(stderr, err) {
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return
	}

	if tryCancelPolicyOrGrants(stderr, err) {
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return
	}

	output.Error(stderr, install.Reason(err))

	if programsReplaced {
		writeUpgradeMigrateGuide(stderr, env, true)
	}
}

func tryCancelPolicyOrGrants(stderr io.Writer, err error) bool {
	var polErr *policy.Error

	//nolint:modernize // preflight wraps policy errors; AsType does not unwrap.
	if errors.As(err, &polErr) {
		cancelPolicyCLI(stderr, err)

		return true
	}

	var grantErr *grants.Error

	//nolint:modernize // preflight wraps grants errors; AsType does not unwrap.
	if errors.As(err, &grantErr) {
		cancelGrantsCLI(stderr, err)

		return true
	}

	if errors.Is(err, policy.ErrMissing) || errors.Is(err, grants.ErrMissing) {
		return false
	}

	return false
}

func writePreflightDiagnosis(stderr io.Writer, err error) bool {
	if note, ok := policy.Corruption(err); ok {
		output.Error(stderr, reasonPolicyUnusable)
		diagnose.Write(stderr, note)

		return true
	}

	if note, ok := grants.Corruption(err); ok {
		output.Error(stderr, reasonMigrateGrantsRead)
		diagnose.Write(stderr, note)

		return true
	}

	if errors.Is(err, policy.ErrMissing) {
		return false
	}

	if polErr, ok := errors.AsType[*policy.Error](err); ok {
		if polErr.Kind == policy.KindPath {
			output.Error(stderr, policy.ReasonPathUnset)

			return true
		}
	}

	return false
}
