//go:build unix

package executor

import (
	"aws-cli-guard/internal/cache"
	"aws-cli-guard/internal/cliversion"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/verbose"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

// PolicyGate decides whether gate.Decide runs before exec.
// When ok is false, a cancellation was written and the process should exit 3.
func PolicyGate(realPath string, env []string, stderr io.Writer) (bool, bool) {
	envLevel, err := verbose.EnvLevel(env)
	if err != nil {
		output.Cancel(stderr, verbose.ReasonInvalid)

		return false, false
	}

	sessionID := session.ID()

	key, err := cliVersionKey(realPath, sessionID)
	if err != nil {
		writeVersionRefuse(stderr, envLevel, env, realPath)
		output.Cancel(stderr, CancellationReason(err))

		return false, false
	}

	result, fromCache := lookupCLIVersion(env, key, stderr)
	if !fromCache {
		result = probeCLIVersion(realPath)
		persistCLIVersion(env, key, result, false)
	}

	maybeWarnCLIVersion(stderr, result, env, key, fromCache)

	if !cliversion.PolicyApplies(result.Relation) {
		writeVersionHandoff(stderr, envLevel, env, realPath)

		return false, true
	}

	return true, true
}

func writeVersionHandoff(stderr io.Writer, level settings.Level, env []string, realPath string) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // version handoff block.
		Home:    homeOriginLabel(env),
		AWS:     realPath,
		Outcome: "version handoff",
	})
}

func writeVersionRefuse(stderr io.Writer, level settings.Level, env []string, realPath string) {
	verbose.Write(stderr, level, verbose.Event{ //nolint:exhaustruct_v5 // version-key refuse block.
		Home:    homeOriginLabel(env),
		AWS:     realPath,
		Outcome: "refuse",
	})
}

func homeOriginLabel(env []string) string {
	label, ok := paths.HomeOrigin(env)
	if !ok {
		return ""
	}

	return label
}

func lookupCLIVersion(
	env []string,
	key cache.CLIVersionKey,
	stderr io.Writer,
) (cliversion.Result, bool) {
	file, err := cache.LoadCLIVersion(env)
	if err != nil {
		if errors.Is(err, cache.ErrCLIVersionUnusable) {
			writeUnusableCacheWarning(stderr, cliVersionCacheFilePath(env))
		}

		return cliversion.Result{Relation: cliversion.RelationUnreadable, Version: ""}, false
	}

	if file.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassCache(), file.Path, file.Mode)
	}

	entry, ok := cache.LookupCLIVersion(file, key)
	if !ok {
		return cliversion.Result{Relation: cliversion.RelationUnreadable, Version: ""}, false
	}

	return cliversion.Result{
		Relation: entry.Relation,
		Version:  entry.CLIVersion,
	}, true
}

func cliVersionCacheFilePath(env []string) string {
	dir, err := paths.CacheDir(env)
	if err != nil {
		return "cli-version.json"
	}

	return filepath.Join(dir, "cli-version.json")
}

func writeUnusableCacheWarning(stderr io.Writer, path string) {
	output.Warning(
		stderr,
		fmt.Sprintf(
			"cache file %q is unusable; continuing. aws-guard recover cache replaces it",
			path,
		),
	)
}

func maybeWarnCLIVersion(
	stderr io.Writer,
	result cliversion.Result,
	env []string,
	key cache.CLIVersionKey,
	fromCache bool,
) {
	if !cliversion.Warns(result.Relation) {
		return
	}

	if fromCache {
		file, err := cache.LoadCLIVersion(env)
		if err != nil {
			output.Warning(stderr, cliversion.WarningText(result))

			return
		}

		entry, ok := cache.LookupCLIVersion(file, key)
		if ok && entry.Warned {
			return
		}
	}

	output.Warning(stderr, cliversion.WarningText(result))
	persistCLIVersion(env, key, result, true)
}

func persistCLIVersion(
	env []string,
	key cache.CLIVersionKey,
	result cliversion.Result,
	warned bool,
) {
	entry := cache.CLIVersionEntry{
		Key:        key,
		Relation:   result.Relation,
		CLIVersion: result.Version,
		Warned:     warned,
	}

	_ = cache.UpsertCLIVersion(env, entry)
}
