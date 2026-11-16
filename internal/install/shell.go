package install

// ShellName identifies a supported interactive shell startup file.
type ShellName string

const (
	// ShellBash selects ~/.bashrc.
	ShellBash ShellName = "bash"
	// ShellZsh selects ~/.zshrc.
	ShellZsh ShellName = "zsh"
)

// ShellOptions configures shell startup integration.
type ShellOptions struct {
	Env    []string
	Home   string
	Prefix string
	Shell  ShellName
	Remove bool
}

// ShellResult reports whether the startup file changed.
type ShellResult struct {
	Changed bool
	Path    string
}
