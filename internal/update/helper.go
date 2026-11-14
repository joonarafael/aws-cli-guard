package update

import (
	"aws-cli-guard/internal/cache"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/version"
	"errors"
	"os"
	"runtime"
	"time"
)

// CheckOptions configures the update-check helper.
type CheckOptions struct {
	Stamp      string
	Executable string
	Env        []string
	Now        time.Time
	Client     ReleaseClient
	Goos       string
	Goarch     string
}

// RunUpdateCheck is the detached helper entrypoint.
//
//nolint:funlen // single orchestration path; splitting would obscure ADR 0030 order.
func RunUpdateCheck(opts CheckOptions) {
	if opts.Stamp == "" {
		opts.Stamp = version.Version
	}

	if !ReleaseStamp(opts.Stamp) {
		return
	}

	prefix, verified := install.VerifiedAWSPrefix(opts.Executable)
	if !verified {
		return
	}

	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}

	if opts.Goos == "" {
		opts.Goos = runtime.GOOS
	}

	if opts.Goarch == "" {
		opts.Goarch = runtime.GOARCH
	}

	env := StripAWSCredentials(opts.Env)

	record, lock, ok := helperBegin(env, opts.Now)
	if !ok {
		return
	}

	defer func() {
		_ = lock.Release()
	}()

	latest, haveRelease, fetchErr := FetchLatest(opts.Client, opts.Stamp)
	if fetchErr != nil {
		_ = writeCheckFailure(env, opts.Stamp, opts.Now, record)

		return
	}

	if !haveRelease {
		_ = writeCheckFailure(env, opts.Stamp, opts.Now, record)

		return
	}

	if !cache.VersionNewer(latest, opts.Stamp) {
		_ = writeChecked(env, opts.Stamp, opts.Now, opts.Stamp, latest, cache.NoticeEmpty)

		return
	}

	setFile, settingsErr := settings.Load(env)
	if settingsErr != nil {
		_ = writeCheckFailure(env, opts.Stamp, opts.Now, record)

		return
	}

	if setFile.Auto != settings.AutoOn {
		_ = writeChecked(env, opts.Stamp, opts.Now, opts.Stamp, latest, cache.NoticeAvailable)

		return
	}

	err := PreflightReplace(prefix, env)
	if err != nil || !SupportedPlatform(opts.Goos, opts.Goarch) {
		_ = writeChecked(env, opts.Stamp, opts.Now, opts.Stamp, latest, cache.NoticeAvailable)

		return
	}

	replaceErr := downloadAndReplace(opts, prefix, latest)
	if replaceErr != nil {
		_ = writeChecked(env, opts.Stamp, opts.Now, opts.Stamp, latest, cache.NoticeAvailable)

		return
	}

	migrateErr := migrateAfterReplace(env, opts.Now)

	notice := cache.NoticeUpdated
	if migrateErr != nil {
		notice = cache.NoticeMigrateFailed
	}

	_ = writeChecked(env, latest, opts.Now, latest, latest, notice)
}

func helperBegin(env []string, now time.Time) (cache.UpdatesRecord, *files.Lock, bool) {
	record, err := cache.LoadUpdates(env)
	if err != nil {
		if errors.Is(err, cache.ErrUpdatesUnusable) {
			return cache.UpdatesRecord{}, nil, false //nolint:exhaustruct_v5 // empty failure.
		}

		return cache.UpdatesRecord{}, nil, false //nolint:exhaustruct_v5 // empty failure.
	}

	if cache.HelperFresh(record, now) {
		return record, nil, false
	}

	err = ensureUpdatesCacheDir(record.Path)
	if err != nil {
		return cache.UpdatesRecord{}, nil, false //nolint:exhaustruct_v5 // empty failure.
	}

	lock, err := files.OpenLock(record.Path)
	if err != nil {
		return cache.UpdatesRecord{}, nil, false //nolint:exhaustruct_v5 // empty failure.
	}

	fresh, err := cache.LoadUpdates(env)
	if err != nil {
		_ = lock.Release()

		return cache.UpdatesRecord{}, nil, false //nolint:exhaustruct_v5 // empty failure.
	}

	if cache.HelperFresh(fresh, now) {
		_ = lock.Release()

		return fresh, nil, false
	}

	return fresh, lock, true
}

//nolint:wrapcheck // cache write errors are ignored by the helper.
//nolint:wrapcheck // cache write errors are ignored by the helper.
func writeCheckFailure(env []string, stamp string, now time.Time, prior cache.UpdatesRecord) error {
	if prior.CheckedAt.IsZero() {
		prior.Current = stamp
		prior.Latest = ""
		prior.Notice = cache.NoticeEmpty
	}

	prior.CheckedAt = now.UTC()

	return cache.WriteUpdates(env, prior)
}

//nolint:wrapcheck // cache write errors are ignored by the helper.
func writeChecked(
	env []string,
	current string,
	now time.Time,
	running, latest, notice string,
) error {
	record := cache.UpdatesRecord{ //nolint:exhaustruct_v5 // Path and Mode set on write.
		CheckedAt: now.UTC(),
		Current:   current,
		Latest:    latest,
		Notice:    notice,
	}

	if record.Current == "" {
		record.Current = running
	}

	return cache.WriteUpdates(env, record)
}

//nolint:wrapcheck // download and replace errors become available notices.
func downloadAndReplace(opts CheckOptions, prefix, latest string) error {
	tempDir, err := os.MkdirTemp("", "aws-cli-guard-update-")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	err = FetchReleaseArchive(
		opts.Client,
		opts.Client.HTTP,
		latest,
		opts.Goos,
		opts.Goarch,
		tempDir,
	)
	if err != nil {
		return err
	}

	return install.ReplaceRelease(prefix, tempDir, latest)
}

//nolint:wrapcheck // migrate errors become migrate-failed notices.
func migrateAfterReplace(env []string, now time.Time) error {
	presentErr := policy.Present(env)
	if presentErr == nil {
		_, err := policy.Migrate(env, now)
		if err != nil {
			return err
		}
	} else if !errors.Is(presentErr, policy.ErrMissing) && !errors.Is(presentErr, os.ErrNotExist) {
		return presentErr
	}

	presentErr = grants.Present(env)
	if presentErr == nil {
		_, err := grants.Migrate(env, now)
		if err != nil {
			return err
		}
	} else if !errors.Is(presentErr, grants.ErrMissing) && !errors.Is(presentErr, os.ErrNotExist) {
		return presentErr
	}

	return nil
}

// RunUpdateCheckFromEnv is the production helper entry using os.Executable.
func RunUpdateCheckFromEnv(env []string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	RunUpdateCheck(CheckOptions{
		Stamp:      version.Version,
		Executable: exe,
		Env:        env,
		Now:        time.Now().UTC(),
		Client:     ReleaseClient{}, //nolint:exhaustruct_v5 // compiled-in URLs.
		Goos:       runtime.GOOS,
		Goarch:     runtime.GOARCH,
	})
}
