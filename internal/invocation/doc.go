// Package invocation parses AWS CLI 2.9.0 top-level global options from argv.
//
// Parse sets PrintsAccessKeys for commands that print access keys per ADR 0013.
// It does not run the AWS CLI. Raising the verified CLI ceiling in ADR 0002
// requires rechecking the option table in this package.
package invocation
