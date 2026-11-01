package cache

import (
	"aws-cli-guard/internal/cliversion"
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	cliVersionFileName    = "cli-version.json"
	cliVersionSchema      = 1
	cliVersionMaxFileSize = 1 << 20
)

// ErrCLIVersionUnusable is returned when cli-version.json is present but cannot be used.
var ErrCLIVersionUnusable = errors.New("aws-cli-guard: cache: cli-version file unusable")

// CLIVersionKey identifies a real AWS CLI binary for one terminal session.
type CLIVersionKey struct {
	Session   string
	Path      string
	Device    uint64
	Inode     uint64
	Size      int64
	MtimeNsec int64
}

// CLIVersionEntry is one cached version check result.
type CLIVersionEntry struct {
	Key        CLIVersionKey
	Relation   cliversion.Relation
	CLIVersion string
	Warned     bool
}

// CLIVersionFile is a loaded cli-version.json.
type CLIVersionFile struct {
	Path    string
	Mode    os.FileMode
	Entries []CLIVersionEntry
}

func cliVersionPath(env []string) (string, error) {
	dir, err := paths.CacheDir(env)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: cache: resolve path: %w", err)
	}

	return filepath.Join(dir, cliVersionFileName), nil
}

// LoadCLIVersion reads cli-version.json from env. ErrCLIVersionUnusable marks corruption.
func LoadCLIVersion(env []string) (CLIVersionFile, error) {
	path, err := cliVersionPath(env)
	if err != nil {
		return CLIVersionFile{}, err
	}

	data, mode, err := files.ReadRegular(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CLIVersionFile{Path: path, Mode: 0, Entries: nil}, nil
		}

		return CLIVersionFile{}, fmt.Errorf("%w: %w", ErrCLIVersionUnusable, err)
	}

	if len(data) > cliVersionMaxFileSize {
		return CLIVersionFile{}, ErrCLIVersionUnusable
	}

	var doc cliVersionDocument

	decodeErr := decode.Decode(decode.JSON, data, &doc, []int{cliVersionSchema})
	if decodeErr != nil {
		return CLIVersionFile{}, fmt.Errorf("%w: %w", ErrCLIVersionUnusable, decodeErr)
	}

	entries, valErr := validateCLIVersionDocument(path, doc)
	if valErr != nil {
		return CLIVersionFile{}, fmt.Errorf("%w: %w", ErrCLIVersionUnusable, valErr)
	}

	return CLIVersionFile{Path: path, Mode: mode, Entries: entries}, nil
}

// LookupCLIVersion returns a matching entry from a loaded file.
func LookupCLIVersion(file CLIVersionFile, key CLIVersionKey) (CLIVersionEntry, bool) {
	for _, entry := range file.Entries {
		if entry.Key == key {
			return entry, true
		}
	}

	return emptyCLIVersionEntry(), false
}

// UpsertCLIVersion merges entry into cli-version.json under env.
func UpsertCLIVersion(env []string, entry CLIVersionEntry) error {
	path, err := cliVersionPath(env)
	if err != nil {
		return err
	}

	current, loadErr := LoadCLIVersion(env)
	if loadErr != nil {
		return loadErr
	}

	list := current.Entries

	list = upsertCLIVersionEntry(list, entry)

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("aws-cli-guard: cache: mkdir: %w", err)
	}

	data, err := encodeCLIVersion(list)
	if err != nil {
		return err
	}

	err = files.WriteAtomic(path, data)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: cache: write cli-version: %w", err)
	}

	return nil
}

func upsertCLIVersionEntry(list []CLIVersionEntry, entry CLIVersionEntry) []CLIVersionEntry {
	out := make([]CLIVersionEntry, 0, len(list)+1)

	replaced := false

	for _, item := range list {
		if item.Key == entry.Key {
			if !replaced {
				out = append(out, entry)
				replaced = true
			}

			continue
		}

		out = append(out, item)
	}

	if !replaced {
		out = append(out, entry)
	}

	return out
}

func encodeCLIVersion(list []CLIVersionEntry) ([]byte, error) {
	doc := cliVersionEncodeDocument{
		Version: cliVersionSchema,
		Entries: make([]cliVersionEncodeEntry, len(list)),
	}

	for i, entry := range list {
		doc.Entries[i] = cliVersionEncodeEntry{
			Session:    entry.Key.Session,
			Path:       entry.Key.Path,
			Device:     entry.Key.Device,
			Inode:      entry.Key.Inode,
			Size:       entry.Key.Size,
			MtimeNsec:  entry.Key.MtimeNsec,
			Relation:   entry.Relation.String(),
			CLIVersion: entry.CLIVersion,
			Warned:     entry.Warned,
		}
	}

	body, err := json.MarshalIndent(&doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: cache: encode cli-version: %w", err)
	}

	body = append(body, '\n')

	return body, nil
}
