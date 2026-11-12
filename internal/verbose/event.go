package verbose

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/settings"
	"io"
)

// Event is one decision-path block. Empty fields are omitted.
type Event struct {
	Home    string
	Subject string
	Command string
	Rule    string
	Grant   string
	Stdin   string
	Caller  string
	AWS     string
	Outcome string
}

// Write prints the decision block when level is verbose.
func Write(writer io.Writer, level settings.Level, event Event) {
	if level != settings.LevelVerbose {
		return
	}

	writeField(writer, "home", event.Home)
	writeField(writer, "subject", event.Subject)
	writeField(writer, "command", event.Command)
	writeField(writer, "rule", event.Rule)
	writeField(writer, "grant", event.Grant)
	writeField(writer, "stdin", event.Stdin)
	writeField(writer, "caller", event.Caller)
	writeField(writer, "aws", event.AWS)
	writeField(writer, "outcome", event.Outcome)
}

func writeField(writer io.Writer, key, value string) {
	if value == "" {
		return
	}

	output.Detail(writer, key+": "+value)
}
