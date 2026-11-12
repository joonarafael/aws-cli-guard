//go:build !unix

package terminal

func readLineOnTerminal(_ string) (string, error) {
	return "", ErrPlatformUnsupported
}
