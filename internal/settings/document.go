package settings

type document struct {
	Version  int               `toml:"version"`
	Identity *identityDocument `toml:"identity"`
	Logging  *loggingDocument  `toml:"logging"`
	Updates  *updatesDocument  `toml:"updates"`
}

type updatesDocument struct {
	Auto *bool `toml:"auto"`
}

type loggingDocument struct {
	Level *string `toml:"level"`
}

type identityDocument struct {
	Verify *bool `toml:"verify"`
}
