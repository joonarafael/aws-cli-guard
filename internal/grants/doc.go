// Package grants loads and writes the system-wide grants file (grants.json).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.GrantsFile and files.ReadRegular, then decodes
// strictly with internal/decode. A missing file means no grants. The wrapper
// entrypoint does not call Load until policy evaluation exists. Corruption builds
// the corrupt-file note that MV-1 and aws-guard recover will print; it does not
// write stderr itself.
package grants
