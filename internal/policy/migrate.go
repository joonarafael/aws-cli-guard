package policy

import (
	"aws-cli-guard/internal/files"
	"errors"
	"fmt"
	"time"
)

// MigrateTargetVersion is the policy schema version migrate rewrites to.
func MigrateTargetVersion() int {
	return readVersionAgentActions
}

// MigrateResult describes one policy migrate run.
type MigrateResult struct {
	Aside       string
	FromVersion int
	Rewrote     bool
}

// Migrate copies an older supported policy aside and rewrites it to the newest enforced version.
func Migrate(env []string, now time.Time) (MigrateResult, error) {
	pol, err := Load(env)
	if err != nil {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, err
	}

	if pol.Version >= readVersionAgentActions {
		return MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}, nil
	}

	aside, err := files.CopyAside(pol.Path, now)
	if err != nil {
		empty := MigrateResult{Aside: "", FromVersion: 0, Rewrote: false}

		return empty, mapMigrateError(pol.Path, err)
	}

	body := encodePolicyV4(pol)

	err = files.WriteAtomic(pol.Path, body)
	if err != nil {
		partial := MigrateResult{Aside: aside, FromVersion: 0, Rewrote: false}

		return partial, mapMigrateError(pol.Path, err)
	}

	return MigrateResult{
		Aside:       aside,
		FromVersion: pol.Version,
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

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: policy: %w", err))
}
