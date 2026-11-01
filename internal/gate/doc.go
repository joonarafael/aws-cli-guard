// Package gate decides whether the aws wrapper may hand off to the real AWS CLI.
//
// Decide recognizes login and returns before loading policy or grants. Otherwise
// it loads policy and grants, applies credential-export and profile rules, and
// writes secret-safe cancellation output to stderr when execution is refused.
package gate
