//go:build !unix

package terminal

func confirmOnTerminal(_ string, _ string) error {
	return ErrPlatformUnsupported
}
