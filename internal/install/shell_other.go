//go:build !unix

package install

import "io"

// ApplyShell edits the shell startup file for Bash or zsh on unix hosts.
func ApplyShell(_ ShellOptions, _ io.Writer) (ShellResult, error) {
	return ShellResult{}, ErrPlatformUnsupported
}
