// Package grants loads and writes the system-wide grants file (grants.json, version 2 on write).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.GrantsFile and files.ReadRegular, then decodes
// strictly with internal/decode. A missing file means no grants. Fingerprint
// fills Source from AWS shared config for a profile name; Equal compares two
// sources; Current finds an unexpired matching grant in a list. aws-guard allow
// calls Fingerprint and Update; aws-guard revoke calls Revoke. The aws wrapper
// loads grants and applies
// Fingerprint and Current through internal/gate. Recover moves a broken grants
// file aside and writes an empty grant list. Corruption builds the corrupt-file
// note that gate prints after output.Cancel; it does not write stderr itself.
package grants
