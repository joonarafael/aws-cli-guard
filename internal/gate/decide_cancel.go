package gate

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"errors"
	"io"
)

func cancelSettingsLoad(stderr io.Writer, err error) {
	if note, ok := settings.Corruption(err); ok {
		output.Cancel(stderr, reasonSettingsUnusable)
		diagnose.Write(stderr, note)

		return
	}

	var setErr *settings.Error

	if errors.As(err, &setErr) && setErr.Kind == settings.KindPath {
		output.Cancel(stderr, reasonSettingsPathUnset)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Cancel(stderr, reasonSettingsPathUnset)

		return
	}

	output.Cancel(stderr, reasonSettingsReadFail)
}

func cancelPolicyLoad(stderr io.Writer, err error) {
	if note, ok := policy.Corruption(err); ok {
		output.Cancel(stderr, reasonPolicyUnusable)
		diagnose.Write(stderr, note)

		return
	}

	if errors.Is(err, policy.ErrMissing) {
		output.Cancel(stderr, policy.ReasonMissing)

		return
	}

	output.Cancel(stderr, reasonPolicyReadFail)
}

func cancelGrantsLoad(stderr io.Writer, err error) {
	if note, ok := grants.Corruption(err); ok {
		output.Cancel(stderr, reasonGrantsUnusable)
		diagnose.Write(stderr, note)

		return
	}

	var grErr *grants.Error

	if errors.As(err, &grErr) && grErr.Kind == grants.KindPath {
		output.Cancel(stderr, reasonGrantsPathUnset)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Cancel(stderr, reasonGrantsPathUnset)

		return
	}

	output.Cancel(stderr, reasonGrantsReadFail)
}
