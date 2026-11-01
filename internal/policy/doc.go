// Package policy loads and validates the user policy file (config.toml).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.ConfigFile and files.ReadRegular, then decodes
// strictly with internal/decode. Evaluate maps a loaded policy and a credential
// subject to allow, prompt, or deny for profile policy. EvaluateExport applies
// [credential_export]; EvaluateGrant applies grant checking for prompt. The wrapper
// The aws wrapper calls Load, Evaluate, and EvaluateExport through internal/gate.
// Recover moves a broken policy file aside and writes the init default. Corruption
// builds the corrupt-file note that gate prints after output.Cancel; it does not
// write stderr itself.
package policy
