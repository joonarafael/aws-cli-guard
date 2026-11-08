package install

// Result describes the outcome of install or upgrade.
type Result struct {
	Upgraded bool
	Prefix   string
}
