// Package settings loads and writes the preferences file (settings.toml).
//
// Callers pass an environment slice; this package does not read os.Environ.
// Create writes the init default when the file is missing. Load reads through
// paths.SettingsFile and files.ReadRegular, then decodes strictly with
// internal/decode. A missing file returns defaults, including logging level
// tight. Optional [logging] level is tight or verbose.
// Optional [updates] auto is absent, off, or on. Absent and off leave the
// updater off; only absent makes init ask.
// The aws wrapper calls Load through internal/gate. Recover moves a broken
// settings file aside and writes the init default. Corruption builds the
// corrupt-file note that gate prints after output.Cancel; it does not write stderr.
package settings
