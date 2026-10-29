package cli

import (
	"aws-cli-guard/internal/cache"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"io"
	"time"
)

const (
	recoverCommand   = "recover"
	recoverArgsExact = 3

	recoverTargetPolicy = "policy"
	recoverTargetGrants = "grants"
	recoverTargetCache  = "cache"

	reasonRecoverUsage       = "usage: aws-guard recover policy|grants|cache"
	reasonRecoverPolicyMiss  = "the policy file does not exist"
	reasonRecoverGrantsMiss  = "the grants file does not exist"
	reasonRecoverPath        = "the file path could not be resolved"
	reasonRecoverConfirm     = "recovery was not confirmed"
	reasonRecoverPolicyWrite = "the policy file could not be written"
	reasonRecoverGrantsWrite = "the grants file could not be written"
	reasonRecoverCacheWrite  = "the cache could not be written"
)

type confirmFunc func(phrase, prompt string) error

func runRecover(args []string, env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	if len(args) != recoverArgsExact {
		output.Error(stderr, reasonRecoverUsage)

		return output.CancelStatus
	}

	target := args[2]

	switch target {
	case recoverTargetPolicy:
		return recoverPolicy(env, stdout, stderr, confirm)
	case recoverTargetGrants:
		return recoverGrants(env, stdout, stderr, confirm)
	case recoverTargetCache:
		return recoverCache(env, stdout, stderr, confirm)
	default:
		output.Error(stderr, reasonRecoverUsage)

		return output.CancelStatus
	}
}

func recoverPolicy(env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	err := policy.Present(env)
	if err != nil {
		output.Error(stderr, recoverPresentReason(err, true))

		return output.CancelStatus
	}

	err = confirm(recoverTargetPolicy, "type policy to replace the policy file with the default")
	if err != nil {
		output.Error(stderr, reasonRecoverConfirm)

		return output.CancelStatus
	}

	aside, err := policy.Recover(env, time.Now())
	if err != nil {
		if aside != "" {
			output.Detail(stdout, "aside: "+aside)
		}

		output.Error(stderr, reasonRecoverPolicyWrite)

		return output.CancelStatus
	}

	output.Detail(stdout, "aside: "+aside)
	output.Detail(stdout, "wrote fresh policy file")

	return 0
}

func recoverGrants(env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	err := grants.Present(env)
	if err != nil {
		output.Error(stderr, recoverPresentReason(err, false))

		return output.CancelStatus
	}

	err = confirm(
		recoverTargetGrants,
		"type grants to replace the grants file with an empty grant list",
	)
	if err != nil {
		output.Error(stderr, reasonRecoverConfirm)

		return output.CancelStatus
	}

	aside, err := grants.Recover(env, time.Now())
	if err != nil {
		if aside != "" {
			output.Detail(stdout, "aside: "+aside)
		}

		output.Error(stderr, reasonRecoverGrantsWrite)

		return output.CancelStatus
	}

	output.Detail(stdout, "aside: "+aside)
	output.Detail(stdout, "wrote fresh grants file")

	return 0
}

func recoverCache(env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	err := cache.Present(env)
	if errors.Is(err, cache.ErrAbsent) {
		return 0
	}

	if err != nil {
		output.Error(stderr, reasonRecoverPath)

		return output.CancelStatus
	}

	err = confirm(recoverTargetCache, "type cache to replace the cache with an empty directory")
	if err != nil {
		output.Error(stderr, reasonRecoverConfirm)

		return output.CancelStatus
	}

	asides, err := cache.Recover(env, time.Now())
	if err != nil {
		for _, aside := range asides {
			output.Detail(stdout, "aside: "+aside)
		}

		output.Error(stderr, reasonRecoverCacheWrite)

		return output.CancelStatus
	}

	for _, aside := range asides {
		output.Detail(stdout, "aside: "+aside)
	}

	output.Detail(stdout, "cache has no live entries")

	return 0
}

func recoverPresentReason(err error, policyFile bool) string {
	if errors.Is(err, paths.ErrHomeUnset) {
		return reasonRecoverPath
	}

	if policyFile {
		if errors.Is(err, policy.ErrMissing) {
			return reasonRecoverPolicyMiss
		}

		return reasonRecoverPath
	}

	if errors.Is(err, grants.ErrMissing) {
		return reasonRecoverGrantsMiss
	}

	return reasonRecoverPath
}
