//go:build !unix

package install

import (
	"io"
)

// Options configures aws-guard install and upgrade.
type Options struct {
	Env           []string
	Prefix        string
	Executable    string
	BeforeUpgrade func() error
}

// Run installs AWS CLI Guard into the chosen prefix.
func Run(_ Options, _ io.Writer) (Result, error) {
	return Result{}, ErrPlatformUnsupported
}

// Upgrade replaces an existing installation in the chosen prefix.
func Upgrade(_ Options, _ io.Writer) (Result, error) {
	return Result{}, ErrPlatformUnsupported
}
