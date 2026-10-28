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

// Source holds the credential-source fingerprint stored with a grant.
type Source struct {
	AccountID         string
	RoleARN           string
	SSOAccountID      string
	SSORoleName       string
	CredentialProcess bool
}
