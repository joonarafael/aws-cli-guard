// Package decode strictly decodes versioned TOML and JSON bytes into caller structs.
//
// Callers pass bytes already read from disk; this package does not open paths.
// The destination struct must carry an exported int field tagged toml:"version" or
// json:"version", matching the format. The same struct type is used for every
// supported schema version; callers pass which integer versions they accept.
//
// Document failures are *Error values with secret-safe messages. Programmer mistakes
// return plain sentinel errors and must not be treated as file corruption.
package decode
