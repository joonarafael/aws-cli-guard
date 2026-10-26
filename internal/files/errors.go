package files

import "errors"

// ErrSymlink is returned when the final path component is a symlink.
var ErrSymlink = errors.New("aws-cli-guard: path is a symlink")

// ErrNotRegular is returned when the path exists but is not a regular file.
var ErrNotRegular = errors.New("aws-cli-guard: path is not a regular file")

// ErrNotDirectory is returned when MkdirPrivate target exists and is not a directory.
var ErrNotDirectory = errors.New("aws-cli-guard: path is not a directory")

// ErrPlatformUnsupported is returned on platforms without a safe-file implementation.
var ErrPlatformUnsupported = errors.New(
	"aws-cli-guard: safe file handling is not supported on this platform yet",
)

// ErrLockAlreadyReleased is returned when Release is called twice on the same Lock.
var ErrLockAlreadyReleased = errors.New("aws-cli-guard: lock already released")

// ErrNoPartialName is returned when no unused partial filename can be chosen.
var ErrNoPartialName = errors.New("aws-cli-guard: no free partial name")
