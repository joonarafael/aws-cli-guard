package grants

import "time"

// Current reports whether list contains a grant for profile that is unexpired at now
// and whose stored source matches current.
func Current(list []Grant, profile string, now time.Time, current Source) bool {
	if profile == "" || len(list) == 0 {
		return false
	}

	for _, grant := range list {
		if grant.Profile != profile {
			continue
		}

		if !now.Before(grant.Expires) {
			continue
		}

		if grant.Source.Equal(current) {
			return true
		}
	}

	return false
}
