package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/terminal"
	"errors"
	"io"
)

const (
	reasonUpdateChoice     = "the update choice was not confirmed"
	reasonSettingsUnusable = "the settings file is unusable"
	reasonSettingsReadFail = "the settings file could not be read"
	updatesInitPrompt      = "automatic updates? type on or off"
	stayOffUpdatesLine     = "automatic updates stay off"
)

type readLineFunc func(prompt string) (string, error)

func headlessUpdatesReadLine(string) (string, error) {
	return "", terminal.ErrNoTerminal
}

type initUpdatesPlan struct {
	headless     bool
	answered     settings.AutoChoice
	printStayOff bool
}

func planInitUpdates(
	_ []string,
	stderr io.Writer,
	readLine readLineFunc,
	loaded settings.File,
	loadOK bool,
	loadErr error,
) (initUpdatesPlan, bool) {
	if !loadOK {
		cancelInitSettings(stderr, loadErr)

		return initUpdatesPlan{}, false //nolint:exhaustruct_v5 // failure plan.
	}

	if loaded.Auto != settings.AutoAbsent {
		return initUpdatesPlan{ //nolint:exhaustruct_v5 // stored answer.
			answered: loaded.Auto,
		}, true
	}

	line, err := readLine(updatesInitPrompt)
	if err != nil {
		return initUpdatesPlan{ //nolint:exhaustruct_v5 // headless init path.
			headless:     true,
			printStayOff: true,
		}, true
	}

	switch line {
	case "on":
		return initUpdatesPlan{ //nolint:exhaustruct_v5 // init answer.
			answered: settings.AutoOn,
		}, true
	case "off":
		return initUpdatesPlan{ //nolint:exhaustruct_v5 // init answer.
			answered: settings.AutoOff,
		}, true
	default:
		output.Error(stderr, reasonUpdateChoice)

		return initUpdatesPlan{}, false //nolint:exhaustruct_v5 // failure plan.
	}
}

func cancelInitSettings(stderr io.Writer, err error) {
	if note, ok := settings.Corruption(err); ok {
		output.Cancel(stderr, reasonSettingsUnusable)
		diagnose.Write(stderr, note)

		return
	}

	var setErr *settings.Error

	if errors.As(err, &setErr) && setErr.Kind == settings.KindPath {
		output.Error(stderr, reasonInitPath)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Error(stderr, reasonInitPath)

		return
	}

	output.Error(stderr, reasonSettingsReadFail)
}

//nolint:wrapcheck // init maps settings write errors through initCreateReason.
func applyInitSettings(
	env []string,
	plan initUpdatesPlan,
) (bool, bool, error) {
	if plan.headless {
		created, err := settings.Create(env)

		return created, false, err
	}

	created, err := settings.WriteAuto(env, plan.answered)
	if err != nil {
		return false, false, err
	}

	if !created {
		return false, true, nil
	}

	return true, false, nil
}

//nolint:wrapcheck // corruption errors must reach cancelInitSettings unchanged.
func loadSettingsForInit(env []string) (settings.File, bool, error) {
	presentErr := settings.Present(env)
	if presentErr != nil {
		var setErr *settings.Error
		if errors.As(presentErr, &setErr) && setErr.Kind == settings.KindMissing {
			file, err := settings.Load(env)
			if err != nil {
				return settings.File{}, false, err
			}

			return file, true, nil
		}

		return settings.File{}, false, presentErr
	}

	file, err := settings.Load(env)
	if err != nil {
		return settings.File{}, false, err
	}

	return file, true, nil
}
