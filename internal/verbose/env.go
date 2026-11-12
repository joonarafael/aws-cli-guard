package verbose

import (
	"aws-cli-guard/internal/environ"
	"aws-cli-guard/internal/settings"
	"errors"
)

const envKey = "AWS_CLI_GUARD_LOG"

// ReasonInvalid is the cancellation reason for a bad AWS_CLI_GUARD_LOG value.
const ReasonInvalid = "AWS_CLI_GUARD_LOG is invalid"

// ErrInvalidEnv is returned when AWS_CLI_GUARD_LOG is not tight or verbose.
var ErrInvalidEnv = errors.New("aws-cli-guard: invalid AWS_CLI_GUARD_LOG")

// EnvLevel parses AWS_CLI_GUARD_LOG. Missing or empty means tight with no error.
func EnvLevel(env []string) (settings.Level, error) {
	raw, ok := environ.Lookup(envKey, env)
	if !ok || raw == "" {
		return settings.LevelTight, nil
	}

	switch settings.Level(raw) {
	case settings.LevelTight, settings.LevelVerbose:
		return settings.Level(raw), nil
	default:
		return settings.LevelTight, ErrInvalidEnv
	}
}

// Effective returns the log level for this invocation after settings load.
// A non-empty env value wins over the file level.
func Effective(env []string, file settings.Level) (settings.Level, error) {
	fromEnv, err := EnvLevel(env)
	if err != nil {
		return settings.LevelTight, err
	}

	raw, ok := environ.Lookup(envKey, env)
	if ok && raw != "" {
		return fromEnv, nil
	}

	return file, nil
}
