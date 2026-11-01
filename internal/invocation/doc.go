// Package invocation parses AWS CLI 2.9.0 top-level global options from argv.
//
// Parse sets PrintsAccessKeys for commands that print access keys per ADR 0013.
// IsLogin recognizes aws sso login for this build. It does not run the AWS CLI.
// Raising the verified CLI ceiling in ADR 0002 requires rechecking the option
// table and login forms in this package.
package invocation
