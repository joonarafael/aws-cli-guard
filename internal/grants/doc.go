// Package grants loads and writes the system-wide grants file (grants.json).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.GrantsFile and files.ReadRegular, then decodes
// strictly with internal/decode. A missing file means no grants. Fingerprint
// fills Source from AWS shared config for a profile name; Equal compares two
// sources. aws-guard allow calls Fingerprint and Update. The aws wrapper
// entrypoint does not call Load, Fingerprint, or Equal until grant checking
// exists (GR-4). Recover moves a broken grants file aside and writes an empty
// grant list. Corruption builds the corrupt-file note that MV-1 will print; it
// does not write stderr itself.
package grants
