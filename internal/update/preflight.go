package update

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/policy"
	"errors"
	"fmt"
	"os"
)

// PreflightReplace returns nil when auto-update may replace the prefix.
//
//nolint:wrapcheck // domain errors are checked with errors.Is at the helper.
func PreflightReplace(prefix string, env []string) error {
	err := install.PrefixUpgradeable(prefix)
	if err != nil {
		return err
	}

	err = policy.Present(env)
	if err != nil {
		if errors.Is(err, policy.ErrMissing) || errors.Is(err, os.ErrNotExist) {
			return preflightGrantsOnly(env)
		}

		return fmt.Errorf("aws-cli-guard: update: policy present: %w", err)
	}

	_, err = policy.Load(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: update: policy load: %w", err)
	}

	return preflightGrantsOnly(env)
}

func preflightGrantsOnly(env []string) error {
	err := grants.Present(env)
	if err != nil {
		if errors.Is(err, grants.ErrMissing) || errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("aws-cli-guard: update: grants present: %w", err)
	}

	_, err = grants.Load(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: update: grants load: %w", err)
	}

	return nil
}
