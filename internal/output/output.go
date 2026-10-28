package output

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// CancelStatus is the process exit code when AWS CLI Guard cancels a command.
const CancelStatus = 3

const (
	errorPrefix   = "aws-cli-guard: error:"
	warningPrefix = "aws-cli-guard: warning:"
	detailPrefix  = "aws-cli-guard:"
)

// Cancel writes the two-line cancellation message to w.
func Cancel(w io.Writer, reason string) {
	safe := stripControls(reason)
	_, _ = fmt.Fprintln(w, errorPrefix, "command was not executed")
	_, _ = fmt.Fprintln(w, errorPrefix, "AWS CLI Guard stopped it:", safe)
}

// Error writes one error line to w for aws-guard control-plane failures.
func Error(w io.Writer, text string) {
	safe := stripControls(text)
	_, _ = fmt.Fprintln(w, errorPrefix, safe)
}

// Warning writes one warning line to w.
func Warning(w io.Writer, text string) {
	safe := stripControls(text)
	_, _ = fmt.Fprintln(w, warningPrefix, safe)
}

// Detail writes one aws-cli-guard detail line to w.
func Detail(w io.Writer, text string) {
	safe := stripControls(text)
	_, _ = fmt.Fprintln(w, detailPrefix, safe)
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
