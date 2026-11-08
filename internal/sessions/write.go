package sessions

import (
	"aws-cli-guard/internal/files"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func writeLocked(path string, list []Record) error {
	valErr := validateRecordList(path, list)
	if valErr != nil {
		return valErr
	}

	data, err := encodeRecords(list)
	if err != nil {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	err = files.WriteAtomic(path, data)
	if err != nil {
		return mapWriteError(path, err)
	}

	return nil
}

func encodeRecords(list []Record) ([]byte, error) {
	if list == nil {
		list = []Record{}
	}

	encoded := encodeDocument{
		Version: writeVersion,
		Records: make([]encodeRec, len(list)),
	}

	for recIndex, rec := range list {
		item := encodeRec{ //nolint:exhaustruct_v5 // TTY set below when present.
			SessionID: rec.SessionID,
			Effect:    string(rec.Effect),
			Token:     rec.Token,
			At:        rec.At.UTC().Format(time.RFC3339),
		}

		if rec.TTY != nil {
			item.TTY = &encodeTTY{
				Path: rec.TTY.Path,
				Dev:  rec.TTY.Dev,
				Ino:  rec.TTY.Ino,
				Rdev: rec.TTY.Rdev,
			}
		}

		encoded.Records[recIndex] = item
	}

	body, err := json.MarshalIndent(&encoded, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: sessions: encode json: %w", err)
	}

	body = append(body, '\n')

	return body, nil
}

func mapWriteError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: sessions: %w", err))
}

func mapReleaseError(path string, err error) error {
	return newError(
		KindWrite,
		path,
		"",
		0,
		0,
		fmt.Errorf("aws-cli-guard: sessions: release lock: %w", err),
	)
}
