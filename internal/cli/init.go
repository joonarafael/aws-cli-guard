package cli

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"errors"
	"io"
)

const (
	initCommand           = "init"
	initArgsExact         = 2
	reasonPolicyExists    = "the policy file already exists"
	reasonInitPath        = "the config path could not be resolved"
	reasonInitPolicyWrite = "the policy file could not be created"
	reasonInitSettings    = "the settings file could not be created"
)

func runInit(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != initArgsExact {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	policyCreated, err := policy.Create(env)
	if err != nil {
		output.Error(stderr, initCreateReason(err, true))

		return output.CancelStatus
	}

	settingsCreated, err := settings.Create(env)
	if err != nil {
		output.Error(stderr, initCreateReason(err, false))

		return output.CancelStatus
	}

	if !policyCreated && !settingsCreated {
		output.Error(stderr, reasonPolicyExists)

		return output.CancelStatus
	}

	writeInitSuccess(stdout, policyCreated, settingsCreated)

	return 0
}

func writeInitSuccess(stdout io.Writer, policyCreated, settingsCreated bool) {
	if policyCreated {
		output.Detail(stdout, "created policy file")
	} else {
		output.Detail(stdout, "policy file left unchanged")
	}

	if settingsCreated {
		output.Detail(stdout, "created settings file")
	} else {
		output.Detail(stdout, "settings file left unchanged")
	}
}

func initCreateReason(err error, policyFile bool) string {
	if errors.Is(err, paths.ErrHomeUnset) {
		return reasonInitPath
	}

	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		if policyFile {
			return reasonInitPolicyWrite
		}

		return reasonInitSettings
	}

	var polErr *policy.Error

	if errors.As(err, &polErr) && polErr.Kind == policy.KindPath {
		return reasonInitPath
	}

	var setErr *settings.Error

	if errors.As(err, &setErr) && setErr.Kind == settings.KindPath {
		return reasonInitPath
	}

	if policyFile {
		return reasonInitPolicyWrite
	}

	return reasonInitSettings
}
