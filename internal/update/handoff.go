package update

import (
	"aws-cli-guard/internal/cache"
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/output"
	"errors"
	"fmt"
	"io"
	"time"
)

// StartFunc starts the detached update-check helper.
type StartFunc func() error

// BeforeHandoff prints at most one cached notice and may start the helper.
func BeforeHandoff(
	stamp, executable string,
	env []string,
	stderr io.Writer,
	now time.Time,
	start StartFunc,
) {
	if !ReleaseStamp(stamp) {
		return
	}

	_, ok := install.VerifiedAWSPrefix(executable)
	if !ok {
		return
	}

	record, err := cache.LoadUpdates(env)
	if err != nil {
		if errors.Is(err, cache.ErrUpdatesUnusable) {
			writeUnusableUpdatesWarning(stderr, record.Path)
		}

		return
	}

	if record.Mode != 0 {
		diagnose.WriteLooseMode(stderr, diagnose.ClassCache(), record.Path, record.Mode)
	}

	if record.Current == stamp && record.Notice != cache.NoticeEmpty {
		tryPrintNotice(env, stderr, record)
	}

	if cache.WrapperStale(record, stamp, now) && start != nil {
		_ = start()
	}
}

func tryPrintNotice(env []string, stderr io.Writer, record cache.UpdatesRecord) {
	lock, err := files.TryOpenLock(record.Path)
	if err != nil || lock == nil {
		return
	}

	defer func() {
		_ = lock.Release()
	}()

	fresh, err := cache.LoadUpdates(env)
	if err != nil {
		return
	}

	if fresh.Notice == cache.NoticeEmpty || fresh.Current != record.Current {
		return
	}

	printNotice(stderr, fresh)
	_ = cache.SaveNoticeClear(env, fresh)
}

func printNotice(stderr io.Writer, record cache.UpdatesRecord) {
	switch record.Notice {
	case cache.NoticeAvailable:
		output.Detail(stderr, fmt.Sprintf("release %s is available", record.Latest))
	case cache.NoticeUpdated:
		output.Detail(stderr, "updated this installation to "+record.Latest)
	case cache.NoticeMigrateFailed:
		output.Warning(stderr, "update replaced the programs; policy and grants were not updated")
		output.Detail(stderr, "fix the file and run aws-guard migrate")
	default:
	}
}

func writeUnusableUpdatesWarning(stderr io.Writer, path string) {
	if path == "" {
		path = "updates.json"
	}

	output.Warning(
		stderr,
		fmt.Sprintf(
			"cache file %q is unusable; continuing. aws-guard recover cache replaces it",
			path,
		),
	)
}
