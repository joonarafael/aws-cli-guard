package session

import (
	"aws-cli-guard/internal/sessions"
	"fmt"
	"os"
)

// SessionOffWarning is printed when policy is skipped for session-off.
const SessionOffWarning = "this terminal session is off and policy is not applied"

// PolicySkipped reports whether the wrapper should hand off without policy for session-off.
func PolicySkipped(file sessions.File, sessionID string) bool {
	latest, ok := sessions.LatestForSession(file, sessionID)
	if !ok {
		return false
	}

	if latest.Effect != sessions.EffectOff {
		return false
	}

	if latest.TTY == nil {
		return false
	}

	if !sessions.ValidTokenPublic(latest.Token) {
		return false
	}

	return ttyMatchesRecord(*latest.TTY)
}

func ttyMatchesRecord(record sessions.TTY) bool {
	if current, ok := CurrentControllingTTY(); ok {
		return ttyEqual(current, record)
	}

	return recordedTTYExists(record)
}

func ttyEqual(a, b sessions.TTY) bool {
	return a.Path == b.Path && a.Dev == b.Dev && a.Ino == b.Ino && a.Rdev == b.Rdev
}

func recordedTTYExists(record sessions.TTY) bool {
	info, err := lstatCharDevice(record.Path)
	if err != nil {
		return false
	}

	return info.Dev == record.Dev && info.Ino == record.Ino && info.Rdev == record.Rdev
}

// CharDeviceStat holds device numbers from lstat.
type CharDeviceStat struct {
	Dev  uint64
	Ino  uint64
	Rdev uint64
}

// LstatCharDeviceStats reads device identity for path. Tests inject alternate behavior.
//
//nolint:gochecknoglobals // test hook for session-off honor without a live tty.
var LstatCharDeviceStats = defaultLstatCharDevice

func defaultLstatCharDevice(path string) (CharDeviceStat, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return CharDeviceStat{}, fmt.Errorf("aws-cli-guard: session: lstat %q: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return CharDeviceStat{}, os.ErrInvalid
	}

	if info.Mode()&os.ModeCharDevice == 0 {
		return CharDeviceStat{}, os.ErrInvalid
	}

	stat, ok := deviceNumbers(info)
	if !ok {
		return CharDeviceStat{}, os.ErrInvalid
	}

	return stat, nil
}

func lstatCharDevice(path string) (CharDeviceStat, error) {
	return LstatCharDeviceStats(path)
}
