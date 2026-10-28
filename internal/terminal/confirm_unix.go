//go:build unix

package terminal

import (
	"os"
)

const devTTY = "/dev/tty"

func confirmOnTerminal(phrase, prompt string) error {
	ttyFile, err := os.OpenFile(devTTY, os.O_RDWR, 0)
	if err != nil {
		return ErrNoTerminal
	}

	defer func() {
		_ = ttyFile.Close()
	}()

	return ConfirmOnReadWriter(phrase, prompt, ttyFile)
}
