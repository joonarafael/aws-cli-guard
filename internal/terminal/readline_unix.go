//go:build unix

package terminal

import (
	"os"
)

func readLineOnTerminal(prompt string) (string, error) {
	ttyFile, err := os.OpenFile(devTTY, os.O_RDWR, 0)
	if err != nil {
		return "", ErrNoTerminal
	}

	defer func() {
		_ = ttyFile.Close()
	}()

	return readLineOnReadWriter(prompt, ttyFile)
}
