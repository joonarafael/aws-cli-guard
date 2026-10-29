package policy

import (
	"aws-cli-guard/internal/paths"
	"errors"
)

// Fixed cancellation reasons for policy.Present failures (no paths or argv).
const (
	ReasonMissing     = "the policy file is missing; run aws-guard init"
	ReasonPathUnset   = "the policy path could not be resolved"
	ReasonPresentFail = "the policy file could not be checked"
)

// PresentReason maps a Present error to a fixed, secret-safe cancellation reason.
func PresentReason(err error) string {
	if errors.Is(err, ErrMissing) {
		return ReasonMissing
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		return ReasonPathUnset
	}

	var polErr *Error

	if errors.As(err, &polErr) && polErr.Kind == KindPath {
		return ReasonPathUnset
	}

	return ReasonPresentFail
}
