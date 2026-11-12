//go:build !unix

package update

import "errors"

var errUnsupported = errors.New("aws-cli-guard: update: unsupported platform")

// StartDetachedCheck is unsupported off unix.
func StartDetachedCheck(_ string, _ []string) error {
	return errUnsupported
}
