//go:build !unix

package install

import (
	"io"
)

// Options configures aws-guard install.
type Options struct {
	Env        []string
	Prefix     string
	Executable string
}

// Run installs AWS CLI Guard into the chosen prefix.
func Run(_ Options, _ io.Writer) error {
	return ErrPlatformUnsupported
}
