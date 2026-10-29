package subject

import (
	"aws-cli-guard/internal/output"
	"io"
)

const defaultProfileName = "default"

const defaultProfileWarning = "using the default profile; a named profile is better practice"

// WarnDefaultProfile writes one stderr warning when subj is the named profile default.
func WarnDefaultProfile(w io.Writer, subj Subject) {
	if subj.Kind != NamedProfile || subj.Name != defaultProfileName {
		return
	}

	output.Warning(w, defaultProfileWarning)
}
