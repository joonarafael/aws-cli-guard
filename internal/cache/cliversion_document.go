package cache

type cliVersionDocument struct {
	Version int                        `json:"version"`
	Entries *[]cliVersionEntryDocument `json:"entries"`
}

type cliVersionEntryDocument struct {
	Session    *string `json:"session"`
	Path       *string `json:"path"`
	Device     *uint64 `json:"device"`
	Inode      *uint64 `json:"inode"`
	Size       *int64  `json:"size"`
	MtimeNsec  *int64  `json:"mtime_nsec"`
	Relation   *string `json:"relation"`
	CLIVersion *string `json:"cli_version"`
	Warned     *bool   `json:"warned"`
}

type cliVersionEncodeDocument struct {
	Version int                     `json:"version"`
	Entries []cliVersionEncodeEntry `json:"entries"`
}

type cliVersionEncodeEntry struct {
	Session    string `json:"session"`
	Path       string `json:"path"`
	Device     uint64 `json:"device"`
	Inode      uint64 `json:"inode"`
	Size       int64  `json:"size"`
	MtimeNsec  int64  `json:"mtime_nsec"`
	Relation   string `json:"relation"`
	CLIVersion string `json:"cli_version"`
	Warned     bool   `json:"warned"`
}
