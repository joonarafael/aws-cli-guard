// Package settings writes the preferences file (settings.toml).
//
// Only aws-guard init writes this file, and nothing reads it. Loading, corrupt-file
// diagnosis, and aws-guard recover settings belong to FS-6.
package settings
