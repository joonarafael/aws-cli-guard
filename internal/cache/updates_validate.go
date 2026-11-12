package cache

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errUpdatesInvalid = errors.New("aws-cli-guard: cache: updates invalid")
	semverPattern     = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

func validateUpdatesDocument(path string, doc updatesDocument) (UpdatesRecord, error) {
	checkedAt, err := requireRFC3339(doc.CheckedAt, "checked_at")
	if err != nil {
		return UpdatesRecord{}, fmt.Errorf("%w in %q: %w", errUpdatesInvalid, path, err)
	}

	current, err := requireSemver(doc.Current, "current")
	if err != nil {
		return UpdatesRecord{}, fmt.Errorf("%w in %q: %w", errUpdatesInvalid, path, err)
	}

	latest := ""
	if doc.Latest != nil && *doc.Latest != "" {
		latest, err = requireSemver(doc.Latest, "latest")
		if err != nil {
			return UpdatesRecord{}, fmt.Errorf("%w in %q: %w", errUpdatesInvalid, path, err)
		}
	}

	notice, err := validateNotice(doc.Notice, current, latest)
	if err != nil {
		return UpdatesRecord{}, fmt.Errorf("%w in %q: %w", errUpdatesInvalid, path, err)
	}

	return UpdatesRecord{
		Path:      path,
		Mode:      0,
		CheckedAt: checkedAt,
		Current:   current,
		Latest:    latest,
		Notice:    notice,
	}, nil
}

func validateNotice(notice *string, current, latest string) (string, error) {
	if notice == nil {
		return "", fmt.Errorf("%w: missing notice", errUpdatesInvalid)
	}

	value := *notice

	switch value {
	case NoticeEmpty, NoticeAvailable, NoticeUpdated, NoticeMigrateFailed:
	default:
		return "", fmt.Errorf("%w: invalid notice", errUpdatesInvalid)
	}

	if value == NoticeAvailable {
		if latest == "" || !VersionNewer(latest, current) {
			return "", fmt.Errorf("%w: available notice", errUpdatesInvalid)
		}
	}

	if value == NoticeUpdated || value == NoticeMigrateFailed {
		if latest == "" || !VersionEqual(latest, current) {
			return "", fmt.Errorf("%w: post-update notice", errUpdatesInvalid)
		}
	}

	return value, nil
}

func requireRFC3339(value *string, key string) (time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return time.Time{}, fmt.Errorf("%w: missing %s", errUpdatesInvalid, key)
	}

	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s", errUpdatesInvalid, key)
	}

	return parsed.UTC(), nil
}

func requireSemver(value *string, key string) (string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "", fmt.Errorf("%w: missing %s", errUpdatesInvalid, key)
	}

	text := strings.TrimSpace(*value)
	if !semverPattern.MatchString(text) {
		return "", fmt.Errorf("%w: %s", errUpdatesInvalid, key)
	}

	return text, nil
}

// VersionNewer reports whether left is numerically greater than right.
func VersionNewer(left, right string) bool {
	return compareVersions(left, right) > 0
}

// VersionEqual reports whether left and right are numerically equal.
func VersionEqual(left, right string) bool {
	return compareVersions(left, right) == 0
}

func compareVersions(left, right string) int {
	leftParts := parseVersionParts(left)
	rightParts := parseVersionParts(right)

	for idx := range 3 {
		if leftParts[idx] > rightParts[idx] {
			return 1
		}

		if leftParts[idx] < rightParts[idx] {
			return -1
		}
	}

	return 0
}

func parseVersionParts(text string) [3]int {
	var out [3]int

	parts := strings.Split(text, ".")
	for idx := range 3 {
		if idx >= len(parts) {
			break
		}

		value, err := strconv.Atoi(parts[idx])
		if err != nil {
			continue
		}

		out[idx] = value
	}

	return out
}

// WrapperStale reports whether the wrapper should start a helper.
func WrapperStale(record UpdatesRecord, running string, now time.Time) bool {
	if record.CheckedAt.IsZero() {
		return true
	}

	if record.CheckedAt.After(now) {
		return true
	}

	if !record.CheckedAt.Add(updatesFreshWindow).After(now) {
		return true
	}

	if record.Current != running {
		return true
	}

	return false
}

// HelperFresh reports whether the helper should skip work.
func HelperFresh(record UpdatesRecord, now time.Time) bool {
	if record.CheckedAt.IsZero() {
		return false
	}

	if record.CheckedAt.After(now) {
		return true
	}

	return record.CheckedAt.Add(updatesFreshWindow).After(now)
}
