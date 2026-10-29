//go:build !unix

package files

import "time"

// MoveAside renames path to a sibling corrupt name.
func MoveAside(_ string, _ time.Time) (string, error) {
	return "", ErrPlatformUnsupported
}
