// Package sessions loads and writes session-off records (sessions.json, version 1 on write).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Load reads through paths.SessionsFile and files.ReadRegular, then decodes
// strictly with internal/decode. A missing file means no records. Append appends
// one record under the lock. aws-guard off and aws-guard on call Append. The aws
// wrapper loads through internal/gate. Recover moves a broken file aside and
// writes an empty record list. Corruption builds the corrupt-file note that gate
// prints after output.Cancel.
package sessions
