// Package invocation parses AWS CLI 2.9.0–2.40.3 top-level global options from argv.
//
// Parse sets PrintsAccessKeys for the fixed set of commands that print access keys.
// IsLogin recognizes aws sso login for this build. It does not run the AWS CLI.
// Raising the verified ceiling in internal/cliversion requires rechecking the option
// table and login forms in this package.
package invocation
