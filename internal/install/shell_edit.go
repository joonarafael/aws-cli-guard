package install

import (
	"errors"
	"fmt"
	"strings"
)

const (
	shellMarkerBegin = "# aws-cli-guard begin"
	shellMarkerEnd   = "# aws-cli-guard end"
)

// ErrShellMarkersConflict is returned when startup marker comments are malformed.
var ErrShellMarkersConflict = errors.New("aws-cli-guard: shell startup markers conflict")

// ErrShellUnsupportedLineEndings is returned when the startup file contains carriage returns.
var ErrShellUnsupportedLineEndings = errors.New(
	"aws-cli-guard: shell startup file has unsupported line endings",
)

func shellBlockLines(prefix string) []string {
	return []string{
		shellMarkerBegin,
		fmt.Sprintf("export PATH='%s:$PATH'", prefix),
		shellMarkerEnd,
	}
}

func shellBlock(prefix string) string {
	return strings.Join(shellBlockLines(prefix), "\n")
}

func markerPositions(content string) (int, int, error) {
	if strings.Contains(content, "\r") {
		return 0, 0, ErrShellUnsupportedLineEndings
	}

	lines := strings.Split(content, "\n")

	var begins, ends []int

	for lineIdx, line := range lines {
		switch line {
		case shellMarkerBegin:
			begins = append(begins, lineIdx)
		case shellMarkerEnd:
			ends = append(ends, lineIdx)
		}
	}

	if len(begins) > 1 || len(ends) > 1 {
		return 0, 0, ErrShellMarkersConflict
	}

	if len(begins) != len(ends) {
		return 0, 0, ErrShellMarkersConflict
	}

	if len(begins) == 0 {
		return -1, -1, nil
	}

	if begins[0] >= ends[0] {
		return 0, 0, ErrShellMarkersConflict
	}

	return begins[0], ends[0], nil
}

func editShellStartup(content string, prefix string, remove bool) (string, bool, error) {
	beginAt, endAt, err := markerPositions(content)
	if err != nil {
		return "", false, err
	}

	if remove {
		if beginAt < 0 {
			return content, false, nil
		}

		lines := strings.Split(content, "\n")
		head := lines[:beginAt]
		tail := lines[endAt+1:]
		updated := make([]string, 0, len(head)+len(tail))
		updated = append(updated, head...)
		updated = append(updated, tail...)

		out := joinStartupLines(updated)

		return out, out != content, nil
	}

	block := shellBlock(prefix)

	if beginAt < 0 {
		return appendShellBlock(content, block)
	}

	lines := strings.Split(content, "\n")

	existing := strings.Join(lines[beginAt:endAt+1], "\n")
	if existing == block {
		return content, false, nil
	}

	head := lines[:beginAt]
	tail := lines[endAt+1:]
	blockLines := shellBlockLines(prefix)
	replaced := make([]string, 0, len(head)+len(blockLines)+len(tail))
	replaced = append(replaced, head...)
	replaced = append(replaced, blockLines...)
	replaced = append(replaced, tail...)

	return joinStartupLines(replaced), true, nil
}

func appendShellBlock(content, block string) (string, bool, error) {
	if content == "" {
		return block + "\n", true, nil
	}

	prefix := content
	if !strings.HasSuffix(prefix, "\n") {
		prefix += "\n"
	}

	next := prefix + block + "\n"

	return next, true, nil
}

func joinStartupLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}

	return out
}
