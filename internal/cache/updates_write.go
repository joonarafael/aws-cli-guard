package cache

import (
	"aws-cli-guard/internal/files"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// WriteUpdates atomically replaces updates.json under env.
func WriteUpdates(env []string, record UpdatesRecord) error {
	path, err := updatesPath(env)
	if err != nil {
		return err
	}

	record.Path = path

	data, err := record.encode()
	if err != nil {
		return err
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("aws-cli-guard: cache: mkdir: %w", err)
	}

	err = files.WriteAtomic(path, data)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: cache: write updates: %w", err)
	}

	return nil
}

// SaveNoticeClear clears notice on an existing record and writes it.
func SaveNoticeClear(env []string, record UpdatesRecord) error {
	record.Notice = NoticeEmpty

	return WriteUpdates(env, record)
}

func encodeJSONDocument(doc any) ([]byte, error) {
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: cache: encode json: %w", err)
	}

	body = append(body, '\n')

	return body, nil
}

// RecordCheckedOnly refreshes checked_at and preserves other fields when possible.
func RecordCheckedOnly(env []string, running string, now time.Time, prior UpdatesRecord) error {
	record := prior
	if record.Current == "" {
		record.Current = running
	}

	record.CheckedAt = now.UTC()

	return WriteUpdates(env, record)
}
