package sessions

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"path/filepath"
	"time"
)

// AppendRequest describes one record to append.
type AppendRequest struct {
	SessionID string
	Effect    Effect
	TTY       *TTY
}

// Append adds one record under lock.
func Append(env []string, req AppendRequest, now time.Time) error {
	if req.Effect == EffectOff && req.TTY == nil {
		return newError(KindInvalid, "", "tty", 0, 0, nil)
	}

	path, err := paths.SessionsFile(env)
	if err != nil {
		return mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return mapWriteError(path, err)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return mapWriteError(path, err)
	}

	writeErr := appendLocked(path, req, now)
	releaseErr := lock.Release()

	if writeErr != nil {
		return writeErr
	}

	if releaseErr != nil {
		return mapReleaseError(path, releaseErr)
	}

	return nil
}

func appendLocked(path string, req AppendRequest, now time.Time) error {
	current, loadErr := loadAtPath(path)
	if loadErr != nil {
		return loadErr
	}

	token, tokErr := generateToken()
	if tokErr != nil {
		return newError(KindWrite, path, "", 0, 0, tokErr)
	}

	rec := Record{
		SessionID: req.SessionID,
		Effect:    req.Effect,
		Token:     token,
		At:        now.UTC().Truncate(time.Second),
		TTY:       req.TTY,
	}

	next := append(append([]Record(nil), current.Records...), rec)

	return writeLocked(path, next)
}
