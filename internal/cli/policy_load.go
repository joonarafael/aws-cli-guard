package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"io"
)

const (
	reasonPolicyRead     = "the policy file could not be read"
	reasonPolicyUnusable = "the policy file is unusable"
)

func loadPolicyForCLI(env []string, stderr io.Writer) (policy.Policy, int) {
	pol, err := policy.Load(env)
	if err != nil {
		cancelPolicyCLI(stderr, err)

		//nolint:exhaustruct_v5 // zero value is intentional on load failure.
		return policy.Policy{}, output.CancelStatus
	}

	diagnose.WriteLooseMode(stderr, diagnose.ClassPolicy(), pol.Path, pol.Mode)

	return pol, 0
}

func cancelPolicyCLI(stderr io.Writer, err error) {
	if note, ok := policy.Corruption(err); ok {
		output.Error(stderr, reasonPolicyUnusable)
		diagnose.Write(stderr, note)

		return
	}

	if errors.Is(err, policy.ErrMissing) {
		output.Error(stderr, policy.ReasonMissing)

		return
	}

	var polErr *policy.Error

	if errors.As(err, &polErr) && polErr.Kind == policy.KindPath {
		output.Error(stderr, policy.ReasonPathUnset)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Error(stderr, policy.ReasonPathUnset)

		return
	}

	output.Error(stderr, reasonPolicyRead)
}
