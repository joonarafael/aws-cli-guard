// Package policy loads and validates the user policy file (config.toml).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.ConfigFile and files.ReadRegular, then decodes
// strictly with internal/decode. The wrapper entrypoint does not call Load
// until policy evaluation exists. Corruption builds the corrupt-file note
// that MV-1 and aws-guard recover will print; it does not write stderr itself.
package policy
