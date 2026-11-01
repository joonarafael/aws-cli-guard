// Package paths resolves AWS CLI Guard on-disk locations.
//
// Guard-owned paths are built from the process user's home in the user database,
// with HOME from the environment slice as a fallback when that lookup fails.
// Test builds with the awsguard_testpaths tag honor AWS_CLI_GUARD_*_DIR overrides.
package paths
