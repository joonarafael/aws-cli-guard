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

func runInit(args []string, env []string, stdout, stderr io.Writer, readLine readLineFunc) int {
	if len(args) != initArgsExact {
		output.Error(stderr, usageLine)

		return output.CancelStatus
	}

	loaded, loadOK, loadErr := loadSettingsForInit(env)

	plan, ok := planInitUpdates(env, stderr, readLine, loaded, loadOK, loadErr)
	if !ok {
		return output.CancelStatus
	}

	policyCreated, err := policy.Create(env)
	if err != nil {
		output.Error(stderr, initCreateReason(err, true))

		return output.CancelStatus
	}

	var settingsCreated, settingsUpdated bool

	settingsWereAbsent := loaded.Auto == settings.AutoAbsent

	switch {
	case plan.headless:
		settingsCreated, err = settings.Create(env)
	case settingsWereAbsent && plan.answered != settings.AutoAbsent:
		settingsCreated, settingsUpdated, err = applyInitSettings(env, plan)
	default:
		settingsCreated, err = settings.Create(env)
	}

	if err != nil {
		output.Error(stderr, initCreateReason(err, false))

		return output.CancelStatus
	}

	if !policyCreated && !settingsCreated {
		if plan.headless {
			writeInitSuccess(stdout, false, false, false, plan.printStayOff)

			return 0
		}

		output.Error(stderr, reasonPolicyExists)

		return output.CancelStatus
	}

	writeInitSuccess(stdout, policyCreated, settingsCreated, settingsUpdated, plan.printStayOff)

	return 0
}

func writeInitSuccess(
	stdout io.Writer,
	policyCreated, settingsCreated, settingsUpdated, printStayOff bool,
) {
	if policyCreated {
		output.Detail(stdout, "created policy file")
	} else {
		output.Detail(stdout, "policy file left unchanged")
	}

	switch {
	case settingsUpdated:
		output.Detail(stdout, "updated settings file")
	case settingsCreated:
		output.Detail(stdout, "created settings file")
	default:
		output.Detail(stdout, "settings file left unchanged")
	}

	if printStayOff {
		output.Detail(stdout, stayOffUpdatesLine)
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
