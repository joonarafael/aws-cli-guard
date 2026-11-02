package settings

type document struct {
	Version  int               `toml:"version"`
	Identity *identityDocument `toml:"identity"`
}

type identityDocument struct {
	Verify *bool `toml:"verify"`
}
