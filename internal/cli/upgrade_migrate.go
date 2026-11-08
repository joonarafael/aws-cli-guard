package cli

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

func preflightUpgradeFiles(env []string) error {
	err := policy.Present(env)
	if err != nil {
		if errors.Is(err, policy.ErrMissing) || errors.Is(err, os.ErrNotExist) {
			return preflightGrantsOnly(env)
		}

		return fmt.Errorf("aws-cli-guard: upgrade preflight policy present: %w", err)
	}

	_, err = policy.Load(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: upgrade preflight policy load: %w", err)
	}

	return preflightGrantsOnly(env)
}

func preflightGrantsOnly(env []string) error {
	err := grants.Present(env)
	if err != nil {
		if errors.Is(err, grants.ErrMissing) || errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("aws-cli-guard: upgrade preflight grants present: %w", err)
	}

	_, err = grants.Load(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: upgrade preflight grants load: %w", err)
	}

	return nil
}

func runMigrateAfterUpgrade(
	env []string,
	stdout, stderr io.Writer,
	programsReplaced bool,
) int {
	policyAbsent, policyFailed := upgradePolicyAbsentOrLoad(env, stderr)
	if policyFailed {
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return output.CancelStatus
	}

	now := time.Now()

	code := upgradeMigratePolicy(env, stdout, stderr, policyAbsent, programsReplaced, now)
	if code != 0 {
		return code
	}

	return upgradeMigrateGrants(env, stdout, stderr, programsReplaced, now)
}

func upgradeMigratePolicy(
	env []string,
	stdout, stderr io.Writer,
	policyAbsent, programsReplaced bool,
	now time.Time,
) int {
	if policyAbsent {
		output.Detail(stdout, "policy: no file to migrate")

		return 0
	}

	pol, err := policy.Load(env)
	if err != nil {
		cancelPolicyCLI(stderr, err)
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return output.CancelStatus
	}

	policyResult, err := policy.Migrate(env, now)
	if err != nil {
		if policyResult.Aside != "" {
			output.Detail(stdout, "aside: "+policyResult.Aside)
		}

		output.Error(stderr, reasonMigratePolicy)
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return output.CancelStatus
	}

	writePolicyMigrateLines(stdout, pol.Version, policyResult)

	return 0
}

func upgradeMigrateGrants(
	env []string,
	stdout, stderr io.Writer,
	programsReplaced bool,
	now time.Time,
) int {
	grantsAbsent, grantsFailed := grantsAbsentOrLoad(env, stderr)
	if grantsFailed {
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return output.CancelStatus
	}

	if grantsAbsent {
		output.Detail(stdout, "grants: no file to migrate")

		return 0
	}

	grantsResult, err := grants.Migrate(env, now)
	if err != nil {
		if grantsResult.Aside != "" {
			output.Detail(stdout, "aside: "+grantsResult.Aside)
		}

		output.Error(stderr, reasonMigrateGrants)
		writeUpgradeMigrateGuide(stderr, env, programsReplaced)

		return output.CancelStatus
	}

	writeGrantsMigrateLines(stdout, grantsResult)

	return 0
}

func upgradePolicyAbsentOrLoad(env []string, stderr io.Writer) (bool, bool) {
	err := policy.Present(env)
	if err != nil {
		if errors.Is(err, policy.ErrMissing) || errors.Is(err, os.ErrNotExist) {
			return true, false
		}

		if errors.Is(err, paths.ErrHomeUnset) {
			output.Error(stderr, policy.ReasonPathUnset)

			return false, true
		}

		cancelPolicyCLI(stderr, err)

		return false, true
	}

	_, err = policy.Load(env)
	if err != nil {
		cancelPolicyCLI(stderr, err)

		return false, true
	}

	return false, false
}
