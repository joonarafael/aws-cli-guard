package settings

import "os"

// AutoChoice is whether automatic updates are enabled in settings.
type AutoChoice int

const (
	// AutoAbsent means the file, table, or key is missing; init may ask and the updater stays off.
	AutoAbsent AutoChoice = iota
	// AutoOff means the user chose off or an explicit false was stored.
	AutoOff
	// AutoOn means the user chose on.
	AutoOn
)

// File is a loaded, validated settings file.
type File struct {
	Path    string
	Version int
	Mode    os.FileMode
	Verify  bool
	Level   Level
	Auto    AutoChoice
}
