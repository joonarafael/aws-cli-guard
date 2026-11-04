package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"errors"
	"io"
	"sort"
	"strconv"
	"time"
)

const (
	grantsCommand        = "grants"
	grantsSubList        = "list"
	grantsMinArgsWithSub = 3
	grantsListArgsExact  = 3

	reasonGrantsUsage    = "usage: aws-guard grants list"
	reasonGrantsPath     = "the file path could not be resolved"
	reasonGrantsUnusable = "the grants file is unusable"
)

func runGrants(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) < grantsMinArgsWithSub {
		output.Error(stderr, reasonGrantsUsage)

		return output.CancelStatus
	}

	switch args[2] {
	case grantsSubList:
		return runGrantsList(args, env, stdout, stderr)
	default:
		output.Error(stderr, reasonGrantsUsage)

		return output.CancelStatus
	}
}

func runGrantsList(args []string, env []string, stdout, stderr io.Writer) int {
	if len(args) != grantsListArgsExact {
		output.Error(stderr, reasonGrantsUsage)

		return output.CancelStatus
	}

	file, status := loadGrantsForCLI(env, stderr)
	if status != 0 {
		return status
	}

	writeGrantsList(stdout, env, file, time.Now())

	return 0
}

func loadGrantsForCLI(env []string, stderr io.Writer) (grants.File, int) {
	file, err := grants.Load(env)
	if err != nil {
		cancelGrantsCLI(stderr, err)

		var empty grants.File

		return empty, output.CancelStatus
	}

	if file.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassGrants(), file.Path, file.Mode)
	}

	return file, 0
}

func cancelGrantsCLI(stderr io.Writer, err error) {
	if note, ok := grants.Corruption(err); ok {
		output.Error(stderr, reasonGrantsUnusable)
		diagnose.Write(stderr, note)

		return
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Error(stderr, reasonGrantsPath)

		return
	}

	if grantErr, ok := errors.AsType[*grants.Error](err); ok {
		switch grantErr.Kind {
		case grants.KindPath:
			output.Error(stderr, reasonGrantsPath)

			return
		case grants.KindMissing,
			grants.KindRead,
			grants.KindDecode,
			grants.KindInvalid,
			grants.KindWrite,
			grants.KindResolve:
			output.Error(stderr, reasonGrantsRead)

			return
		}
	}

	output.Error(stderr, reasonGrantsRead)
}

func writeGrantsList(stdout io.Writer, env []string, file grants.File, now time.Time) {
	if len(file.Grants) == 0 {
		output.Detail(stdout, "no grants")

		return
	}

	output.Detail(stdout, "version: "+strconv.Itoa(file.Version))

	sorted := append([]grants.Grant(nil), file.Grants...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := sorted[i], sorted[j]
		if left.Profile != right.Profile {
			return left.Profile < right.Profile
		}

		return left.Expires.Before(right.Expires)
	})

	liveByProfile := liveFingerprints(env, sorted, now)

	for _, grant := range sorted {
		writeOneGrantList(stdout, grant, now, liveByProfile[grant.Profile])
	}
}

func liveFingerprints(env []string, list []grants.Grant, now time.Time) map[string]liveFingerprint {
	out := make(map[string]liveFingerprint)

	for _, grant := range list {
		if !now.Before(grant.Expires) {
			continue
		}

		if _, ok := out[grant.Profile]; ok {
			continue
		}

		live, err := grants.Fingerprint(env, grant.Profile)
		out[grant.Profile] = liveFingerprint{source: live, err: err}
	}

	return out
}

type liveFingerprint struct {
	source grants.Source
	err    error
}

func writeOneGrantList(stdout io.Writer, grant grants.Grant, now time.Time, live liveFingerprint) {
	output.Detail(stdout, "profile: "+grant.Profile)
	output.Detail(stdout, "expires: "+grant.Expires.UTC().Format(time.RFC3339))

	status := grants.GrantStatus(grant, now, live.source, live.err)
	output.Detail(stdout, "grant: "+status.Label())

	for _, line := range grants.IdentityLines(grant.Source) {
		output.Detail(stdout, line)
	}
}
