package grants

import (
	"os"
	"time"
)

// File is a loaded grants file.
type File struct {
	Path    string
	Version int
	Mode    os.FileMode
	Grants  []Grant
}

// Grant is one profile approval on this installation.
type Grant struct {
	Profile string
	Expires time.Time
	Source  Source
}

// ChainHop is one upstream profile snapshot in a version 2 source fingerprint.
type ChainHop struct {
	Profile           string
	AccountID         string
	RoleARN           string
	SSOAccountID      string
	SSORoleName       string
	CredentialProcess bool
	SourceProfile     string
	CredentialSource  string
	SSOSession        string
	SSOStartURL       string
	SSORegion         string
	ExternalID        string
}

// Source holds the credential-source fingerprint stored with a grant.
type Source struct {
	AccountID         string
	RoleARN           string
	SSOAccountID      string
	SSORoleName       string
	CredentialProcess bool
	SourceProfile     string
	CredentialSource  string
	SSOSession        string
	SSOStartURL       string
	SSORegion         string
	ExternalID        string
	Chain             []ChainHop
}

func emptyChainHop(profile string) ChainHop {
	return ChainHop{
		Profile:           profile,
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
		SourceProfile:     "",
		CredentialSource:  "",
		SSOSession:        "",
		SSOStartURL:       "",
		SSORegion:         "",
		ExternalID:        "",
	}
}
