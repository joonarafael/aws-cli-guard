package terminal

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	promptPrefix = "aws-cli-guard: "
	maxLineBytes = 1024
)

// Confirm reads one line from the controlling terminal and compares it to phrase.
func Confirm(phrase, prompt string) error {
	err := validatePhrase(phrase)
	if err != nil {
		return err
	}

	return confirmOnTerminal(phrase, prompt)
}

// ConfirmOnReadWriter runs confirmation on an arbitrary ReadWriter (tests and /dev/tty).
func ConfirmOnReadWriter(phrase, prompt string, tty io.ReadWriter) error {
	return confirmOnReadWriter(phrase, prompt, tty)
}

func validatePhrase(phrase string) error {
	if phrase == "" {
		return ErrInvalidPhrase
	}

	for _, r := range phrase {
		if r == unicode.ReplacementChar {
			continue
		}

		if unicode.IsControl(r) {
			return ErrInvalidPhrase
		}
	}

	return nil
}

func confirmOnReadWriter(phrase, prompt string, tty io.ReadWriter) error {
	err := validatePhrase(phrase)
	if err != nil {
		return err
	}

	safePrompt := stripControls(prompt)

	_, writeErr := fmt.Fprintln(tty, promptPrefix+safePrompt)
	if writeErr != nil {
		return ErrNoTerminal
	}

	reader := bufio.NewReader(tty)

	typed, err := readConfirmationLine(reader)
	if err != nil {
		return ErrMismatch
	}

	if typed != phrase {
		return ErrMismatch
	}

	return nil
}

func readConfirmationLine(reader *bufio.Reader) (string, error) {
	var buf []byte

	for {
		next, err := reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("read confirmation byte: %w", err)
		}

		if next == '\n' {
			break
		}

		if len(buf) >= maxLineBytes {
			return "", ErrMismatch
		}

		buf = append(buf, next)
	}

	typed := string(buf)
	if len(typed) > 0 && typed[len(typed)-1] == '\r' {
		typed = typed[:len(typed)-1]
	}

	return typed, nil
}

func stripControls(text string) string {
	return strings.Map(func(runeVal rune) rune {
		if runeVal == unicode.ReplacementChar {
			return runeVal
		}

		if unicode.IsControl(runeVal) {
			return -1
		}

		return runeVal
	}, text)
}
