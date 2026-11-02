package gate

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/subject"
	"io"
)

const identityUnverifiedLine = "identity: configured source only; the caller was not verified"

func writeProfileIdentityDetails(stderr io.Writer, env []string, subj subject.Subject) {
	if subj.Kind != subject.NamedProfile || subj.Name == "" {
		return
	}

	source, err := grants.Fingerprint(env, subj.Name)
	if err != nil {
		return
	}

	writeAccountDetails(stderr, source)
	writeRoleDetails(stderr, source)

	if sourceNeedsUnverifiedLine(source) {
		output.Detail(stderr, identityUnverifiedLine)
	}
}

func writeAccountDetails(stderr io.Writer, source grants.Source) {
	if source.AccountID != "" {
		output.Detail(stderr, "account: "+source.AccountID)
	}

	if source.SSOAccountID != "" {
		if source.AccountID != "" {
			output.Detail(stderr, "sso account: "+source.SSOAccountID)

			return
		}

		output.Detail(stderr, "account: "+source.SSOAccountID)
	}
}

func writeRoleDetails(stderr io.Writer, source grants.Source) {
	if source.RoleARN != "" {
		output.Detail(stderr, "role: "+source.RoleARN)
	}

	if source.SSORoleName != "" {
		if source.RoleARN != "" {
			output.Detail(stderr, "sso role: "+source.SSORoleName)

			return
		}

		output.Detail(stderr, "role: "+source.SSORoleName)
	}
}

func sourceNeedsUnverifiedLine(source grants.Source) bool {
	return source.CredentialProcess ||
		source.CredentialSource != "" ||
		source.SourceProfile != ""
}
