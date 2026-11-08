package sessions

import (
	"os"
	"time"
)

// Effect is whether a record turns session policy off or on.
type Effect string

const (
	// EffectOff skips policy for matching callers in the session.
	EffectOff Effect = "off"
	// EffectOn restores policy for the session.
	EffectOn Effect = "on"
)

// TTY identifies the controlling terminal captured at confirmation.
type TTY struct {
	Path string
	Dev  uint64
	Ino  uint64
	Rdev uint64
}

// Record is one append-only session-off event.
type Record struct {
	SessionID string
	Effect    Effect
	Token     string
	At        time.Time
	TTY       *TTY
}

// File is a loaded sessions file.
type File struct {
	Path    string
	Version int
	Mode    os.FileMode
	Records []Record
}
