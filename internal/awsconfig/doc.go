// Package awsconfig reads the AWS CLI shared config and credentials files.
//
// Callers pass an environment slice; this package does not read os.Environ.
// Paths honor AWS_CONFIG_FILE and AWS_SHARED_CREDENTIALS_FILE, then default
// locations under HOME. Missing files are treated as empty. Unreadable paths
// that exist are errors so later grant logic cannot treat them as absent.
//
// Parsing follows botocore configloader (profile sections, credentials merge,
// one-level nested blocks). Duplicate keys or sections are errors, matching
// strict ConfigParser behavior. Secret values and credential_process commands
// are not returned; only presence flags are kept.
package awsconfig
