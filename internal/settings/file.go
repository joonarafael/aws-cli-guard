package settings

import "os"

// File is a loaded, validated settings file.
type File struct {
	Path    string
	Version int
	Mode    os.FileMode
	Verify  bool
	Level   Level
}
