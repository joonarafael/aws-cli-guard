// Package policy loads and validates the user policy file (config.toml).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.ConfigFile and files.ReadRegular, then decodes
// strictly with internal/decode. Evaluate maps a loaded policy and a credential
// subject to allow, prompt, or deny for profile policy. It does not check grants
// or credential_export. The wrapper calls Present before handoff; it does not
// call Load or Evaluate until MV-1. Corruption builds the corrupt-file note that
// MV-1 and aws-guard recover will print; it does not write stderr itself.
package policy
