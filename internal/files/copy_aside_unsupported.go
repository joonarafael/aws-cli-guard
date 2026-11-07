//go:build !unix

package files

import "time"

// CopyAside copies path to a sibling migrated name.
func CopyAside(_ string, _ time.Time) (string, error) {
	return "", ErrPlatformUnsupported
}
