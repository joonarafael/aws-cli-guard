package grants

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MigrateTargetVersion is the grants schema version migrate rewrites to.
func MigrateTargetVersion() int {
	return writeVersion
}

// MigrateResult describes one grants migrate run.
type MigrateResult struct {
	Aside       string
	FromVersion int
	Rewrote     bool
}

// Migrate copies a version 1 grants file aside and rewrites it to version 2.
func Migrate(env []string, now time.Time) (MigrateResult, error) {
	path, err := paths.GrantsFile(env)
	if err != nil {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, mapPathError(err)
	}

	_, statErr := os.Lstat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, newError(
				KindMissing,
				path,
				"",
				0,
				0,
				fmt.Errorf("%w: %w", ErrMissing, os.ErrNotExist),
			)
		}

		empty := MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}

		return empty, mapMigrateError(path, statErr)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, mapMigrateError(path, err)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, mapMigrateError(path, err)
	}

	result, migrateErr := migrateLocked(path, now)
	releaseErr := lock.Release()

	if migrateErr != nil {
		return result, migrateErr
	}

	if releaseErr != nil {
		return result, mapReleaseError(path, releaseErr)
	}

	return result, nil
}

func migrateLocked(path string, now time.Time) (MigrateResult, error) {
	current, loadErr := loadAtPath(path)
	if loadErr != nil {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, loadErr
	}

	if current.Version >= writeVersion {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, nil
	}

	aside, err := files.CopyAside(path, now)
	if err != nil {
		failed := MigrateResult{Aside: "", FromVersion: current.Version, Rewrote: false}

		return failed, mapMigrateError(path, err)
	}

	writeErr := writeLocked(path, current.Grants)
	if writeErr != nil {
		return MigrateResult{Aside: aside, FromVersion: 0, Rewrote: false}, writeErr
	}

	return MigrateResult{
		Aside:       aside,
		FromVersion: current.Version,
		Rewrote:     true,
	}, nil
}

func mapMigrateError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	if _, ok := errors.AsType[*Error](err); ok {
		return err
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: grants: %w", err))
}
