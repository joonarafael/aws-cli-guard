package cache

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	updatesFileName    = "updates.json"
	updatesSchema      = 1
	updatesMaxFileSize = 1 << 20
	updatesFreshWindow = 24 * time.Hour
)

// ErrUpdatesUnusable is returned when updates.json is present but cannot be used.
var ErrUpdatesUnusable = errors.New("aws-cli-guard: cache: updates file unusable")

// Notice values in updates.json.
const (
	NoticeEmpty         = ""
	NoticeAvailable     = "available"
	NoticeUpdated       = "updated"
	NoticeMigrateFailed = "migrate-failed"
)

// UpdatesRecord is a loaded updates.json.
type UpdatesRecord struct {
	Path      string
	Mode      os.FileMode
	CheckedAt time.Time
	Current   string
	Latest    string
	Notice    string
}

func updatesPath(env []string) (string, error) {
	dir, err := paths.CacheDir(env)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: cache: resolve path: %w", err)
	}

	return filepath.Join(dir, updatesFileName), nil
}

// LoadUpdates reads updates.json from env.
func LoadUpdates(env []string) (UpdatesRecord, error) {
	path, err := updatesPath(env)
	if err != nil {
		return UpdatesRecord{}, err
	}

	data, mode, err := files.ReadRegular(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UpdatesRecord{Path: path, Mode: 0}, nil //nolint:exhaustruct_v5 // missing file.
		}

		return UpdatesRecord{}, fmt.Errorf("%w: %w", ErrUpdatesUnusable, err)
	}

	if len(data) > updatesMaxFileSize {
		return UpdatesRecord{}, ErrUpdatesUnusable
	}

	var doc updatesDocument

	decodeErr := decode.Decode(decode.JSON, data, &doc, []int{updatesSchema})
	if decodeErr != nil {
		return UpdatesRecord{}, fmt.Errorf("%w: %w", ErrUpdatesUnusable, decodeErr)
	}

	record, valErr := validateUpdatesDocument(path, doc)
	if valErr != nil {
		return UpdatesRecord{}, fmt.Errorf("%w: %w", ErrUpdatesUnusable, valErr)
	}

	record.Path = path
	record.Mode = mode

	return record, nil
}
