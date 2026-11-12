// Package gate decides whether the aws wrapper may hand off to the real AWS CLI.
//
// Decide honors AWS_CLI_GUARD_LOG and settings.toml [logging] level for verbose
// decision-path lines on stderr. It recognizes login and returns before loading
// settings, policy, or grants. Otherwise it loads settings, policy, and grants,
// classifies stdin, applies optional environment-marker recognition, agent profile
// rules for recognized callers, credential-export and profile rules, and writes
// secret-safe cancellation output to stderr when execution is refused.
package gate
