package terminal

import (
	"bufio"
	"fmt"
	"io"
)

// ReadLine reads one line from the controlling terminal after writing prompt.
func ReadLine(prompt string) (string, error) {
	return readLineOnTerminal(prompt)
}

// ReadLineOnReadWriter reads one line after writing prompt to tty.
func ReadLineOnReadWriter(prompt string, tty io.ReadWriter) (string, error) {
	return readLineOnReadWriter(prompt, tty)
}

func readLineOnReadWriter(prompt string, tty io.ReadWriter) (string, error) {
	safePrompt := stripControls(prompt)

	_, writeErr := fmt.Fprintln(tty, promptPrefix+safePrompt)
	if writeErr != nil {
		return "", ErrNoTerminal
	}

	reader := bufio.NewReader(tty)

	typed, err := readConfirmationLine(reader)
	if err != nil {
		return "", ErrMismatch
	}

	return typed, nil
}
