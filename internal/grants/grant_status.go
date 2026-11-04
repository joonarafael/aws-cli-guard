package grants

import "time"

// GrantStatusKind classifies one stored grant for display.
type GrantStatusKind int

// Grant status values for list output.
const (
	GrantStatusCurrent GrantStatusKind = iota + 1
	GrantStatusExpired
	GrantStatusSourceMismatch
	GrantStatusSourceUnread
)

// Label returns the grant status line value for CLI output.
func (k GrantStatusKind) Label() string {
	switch k {
	case GrantStatusCurrent:
		return "current"
	case GrantStatusExpired:
		return "expired"
	case GrantStatusSourceMismatch:
		return "source mismatch"
	case GrantStatusSourceUnread:
		return "source unread"
	default:
		return "unknown"
	}
}

// GrantStatus classifies grant at now against live and readErr from Fingerprint.
func GrantStatus(grant Grant, now time.Time, live Source, readErr error) GrantStatusKind {
	if !now.Before(grant.Expires) {
		return GrantStatusExpired
	}

	if readErr != nil {
		return GrantStatusSourceUnread
	}

	if !grant.Source.Equal(live) {
		return GrantStatusSourceMismatch
	}

	return GrantStatusCurrent
}
