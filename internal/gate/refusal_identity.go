package gate

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/subject"
	"io"
)

func writeProfileIdentityDetails(stderr io.Writer, env []string, subj subject.Subject) {
	if subj.Kind != subject.NamedProfile || subj.Name == "" {
		return
	}

	source, err := grants.Fingerprint(env, subj.Name)
	if err != nil {
		return
	}

	for _, line := range grants.IdentityLines(source) {
		output.Detail(stderr, line)
	}
}
