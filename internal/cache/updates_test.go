package cache_test

import (
	"aws-cli-guard/internal/cache"
	"testing"
	"time"
)

func TestVersionCompareLeadingZeros(t *testing.T) {
	t.Parallel()

	if !cache.VersionEqual("1.02.3", "1.2.3") {
		t.Fatal("expected equal versions")
	}

	if !cache.VersionNewer("1.3.0", "1.2.3") {
		t.Fatal("expected newer")
	}
}

func TestWrapperStaleMissingCheckedAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	record := cache.UpdatesRecord{Current: "1.0.0"} //nolint:exhaustruct_v5 // partial record.

	if !cache.WrapperStale(record, "1.0.0", now) {
		t.Fatal("missing checked_at should be stale")
	}
}

func TestHelperFreshWithinWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	record := cache.UpdatesRecord{ //nolint:exhaustruct_v5 // partial record.
		CheckedAt: now.Add(-time.Hour),
		Current:   "1.0.0",
	}

	if !cache.HelperFresh(record, now) {
		t.Fatal("expected fresh within 24h")
	}
}
