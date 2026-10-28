package diagnose

import (
	"aws-cli-guard/internal/output"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Write prints a corrupt-file diagnosis to w.
func Write(writer io.Writer, note Note) {
	className := note.Class.Name
	if className == "" {
		output.Detail(writer, "file: unknown")
	} else {
		output.Detail(writer, "file: "+className)
	}

	output.Detail(writer, "path: "+note.Path)
	output.Detail(writer, "failure: "+note.Failure)

	if note.Line > 0 {
		output.Detail(writer, fmt.Sprintf("location: line %d column %d", note.Line, note.Column))
	}

	if note.Key != "" {
		output.Detail(writer, "key: "+note.Key)
	}

	if note.HasVersion {
		output.Detail(writer, "version: "+strconv.Itoa(note.Version))
	}

	if len(note.Supported) > 0 {
		output.Detail(writer, "supported versions: "+formatIntList(note.Supported))
	}

	output.Detail(writer, "the file was not modified")
	output.Detail(writer, "fix the file in place, or replace it with a fresh one")

	if note.Class.Recover != "" {
		output.Detail(writer, note.Class.Recover)
	}
}

// WriteLooseMode prints a warning when mode is looser than 0600.
func WriteLooseMode(writer io.Writer, class Class, path string, mode os.FileMode) {
	if mode.Perm()&^0o600 == 0 {
		return
	}

	name := class.Name
	if name == "" {
		name = "file"
	}

	output.Warning(
		writer,
		fmt.Sprintf("%s file %q has loose permissions (%04o)", name, path, mode.Perm()),
	)
}

func formatIntList(vals []int) string {
	parts := make([]string, len(vals))
	for i, val := range vals {
		parts[i] = strconv.Itoa(val)
	}

	return strings.Join(parts, ", ")
}
