package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	migrateCommand   = "migrate"
	migrateArgsExact = 2

	reasonMigrateUsage      = "usage: aws-guard migrate"
	reasonMigratePolicy     = "the policy file could not be migrated"
	reasonMigrateGrants     = "the grants file could not be migrated"
	reasonMigrateGrantsRead = "the grants file could not be read"
)

func runMigrate(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != migrateArgsExact {
		output.Error(stderr, reasonMigrateUsage)

		return output.CancelStatus
	}

	pol, err := policy.Load(env)
	if err != nil {
		cancelPolicyCLI(stderr, err)

		return output.CancelStatus
	}

	diagnose.WriteLooseMode(stderr, diagnose.ClassPolicy(), pol.Path, pol.Mode)

	grantsAbsent, grantsFailed := grantsAbsentOrLoad(env, stderr)
	if grantsFailed {
		return output.CancelStatus
	}

	now := time.Now()

	policyResult, err := policy.Migrate(env, now)
	if err != nil {
		if policyResult.Aside != "" {
			output.Detail(stdout, "aside: "+policyResult.Aside)
		}

		output.Error(stderr, reasonMigratePolicy)

		return output.CancelStatus
	}

	writePolicyMigrateLines(stdout, pol.Version, policyResult)

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

		return output.CancelStatus
	}

	writeGrantsMigrateLines(stdout, grantsResult)

	return 0
}

func grantsAbsentOrLoad(env []string, stderr io.Writer) (bool, bool) {
	err := grants.Present(env)
	if err != nil {
		if errors.Is(err, grants.ErrMissing) {
			return true, false
		}

		cancelGrantsPresent(stderr, err)

		return false, true
	}

	_, err = grants.Load(env)
	if err != nil {
		cancelGrantsCLI(stderr, err)

		return false, true
	}

	return false, false
}

func writePolicyMigrateLines(stdout io.Writer, preVersion int, result policy.MigrateResult) {
	if result.Rewrote {
		output.Detail(
			stdout,
			fmt.Sprintf(
				"policy: rewrote version %d to version %d",
				result.FromVersion,
				policy.MigrateTargetVersion(),
			),
		)
		output.Detail(stdout, "aside: "+result.Aside)

		return
	}

	if preVersion >= policy.MigrateTargetVersion() {
		output.Detail(stdout, "policy: already version 4")
	}
}

func writeGrantsMigrateLines(stdout io.Writer, result grants.MigrateResult) {
	if result.Rewrote {
		output.Detail(
			stdout,
			fmt.Sprintf(
				"grants: rewrote version %d to version %d",
				result.FromVersion,
				grants.MigrateTargetVersion(),
			),
		)
		output.Detail(stdout, "aside: "+result.Aside)

		return
	}

	output.Detail(stdout, "grants: already version 2")
}

func cancelGrantsPresent(stderr io.Writer, err error) {
	if errors.Is(err, paths.ErrHomeUnset) {
		output.Error(stderr, reasonGrantsPath)

		return
	}

	var grantErr *grants.Error

	if errors.As(err, &grantErr) && grantErr.Kind == grants.KindPath {
		output.Error(stderr, reasonGrantsPath)

		return
	}

	output.Error(stderr, reasonMigrateGrantsRead)
}
