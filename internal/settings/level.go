package settings

// Level is the operational log verbosity in settings.toml [logging] or AWS_CLI_GUARD_LOG.
type Level string

const (
	// LevelTight is the default: no decision block on an ordinary allow.
	LevelTight Level = "tight"
	// LevelVerbose adds the decision block on stderr.
	LevelVerbose Level = "verbose"
)
