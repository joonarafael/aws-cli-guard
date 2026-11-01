package cache

import (
	"aws-cli-guard/internal/cliversion"
	"errors"
	"fmt"
	"strconv"
)

var errCLIVersionInvalid = errors.New("aws-cli-guard: cache: cli-version invalid")

func validateCLIVersionDocument(path string, doc cliVersionDocument) ([]CLIVersionEntry, error) {
	if doc.Entries == nil {
		return nil, fmt.Errorf("%w: missing entries", errCLIVersionInvalid)
	}

	out := make([]CLIVersionEntry, 0, len(*doc.Entries))
	seen := make(map[CLIVersionKey]struct{}, len(*doc.Entries))

	for i, item := range *doc.Entries {
		prefix := "entries." + strconv.Itoa(i)

		entry, err := validateCLIVersionEntry(item, prefix)
		if err != nil {
			return nil, fmt.Errorf("%w in %q: %w", errCLIVersionInvalid, path, err)
		}

		if _, dup := seen[entry.Key]; dup {
			return nil, fmt.Errorf("%w in %q: duplicate entry key", errCLIVersionInvalid, path)
		}

		seen[entry.Key] = struct{}{}
		out = append(out, entry)
	}

	return out, nil
}

func validateCLIVersionEntry(item cliVersionEntryDocument, prefix string) (CLIVersionEntry, error) {
	key, err := validateCLIVersionKey(item, prefix)
	if err != nil {
		return emptyCLIVersionEntry(), err
	}

	rel, cliVer, err := validateCLIVersionRelation(item, prefix)
	if err != nil {
		return emptyCLIVersionEntry(), err
	}

	return CLIVersionEntry{
		Key:        key,
		Relation:   rel,
		CLIVersion: cliVer,
		Warned:     *item.Warned,
	}, nil
}

func validateCLIVersionKey(item cliVersionEntryDocument, prefix string) (CLIVersionKey, error) {
	session, err := requireString(item.Session, prefix+".session")
	if err != nil {
		return CLIVersionKey{}, err
	}

	pathVal, err := requireString(item.Path, prefix+".path")
	if err != nil {
		return CLIVersionKey{}, err
	}

	device, err := requireUint64(item.Device, prefix+".device")
	if err != nil {
		return CLIVersionKey{}, err
	}

	inode, err := requireUint64(item.Inode, prefix+".inode")
	if err != nil {
		return CLIVersionKey{}, err
	}

	size, err := requireInt64(item.Size, prefix+".size")
	if err != nil {
		return CLIVersionKey{}, err
	}

	mtime, err := requireInt64(item.MtimeNsec, prefix+".mtime_nsec")
	if err != nil {
		return CLIVersionKey{}, err
	}

	if item.Warned == nil {
		return CLIVersionKey{}, fieldMissing(prefix + ".warned")
	}

	return CLIVersionKey{
		Session:   session,
		Path:      pathVal,
		Device:    device,
		Inode:     inode,
		Size:      size,
		MtimeNsec: mtime,
	}, nil
}

func validateCLIVersionRelation(
	item cliVersionEntryDocument,
	prefix string,
) (cliversion.Relation, string, error) {
	if item.Relation == nil {
		return cliversion.RelationUnreadable, "", fieldMissing(prefix + ".relation")
	}

	if item.CLIVersion == nil {
		return cliversion.RelationUnreadable, "", fieldMissing(prefix + ".cli_version")
	}

	rel, ok := cliversion.ParseRelation(*item.Relation)
	if !ok {
		return cliversion.RelationUnreadable, "", fieldInvalid(prefix + ".relation")
	}

	cliVer := *item.CLIVersion
	if rel == cliversion.RelationUnreadable {
		if cliVer != "" {
			return cliversion.RelationUnreadable, "", fieldInvalid(prefix + ".cli_version")
		}
	} else if cliVer == "" {
		return cliversion.RelationUnreadable, "", fieldInvalid(prefix + ".cli_version")
	}

	return rel, cliVer, nil
}

func emptyCLIVersionEntry() CLIVersionEntry {
	return CLIVersionEntry{
		Key: CLIVersionKey{
			Session:   "",
			Path:      "",
			Device:    0,
			Inode:     0,
			Size:      0,
			MtimeNsec: 0,
		},
		Relation:   cliversion.RelationUnreadable,
		CLIVersion: "",
		Warned:     false,
	}
}

func requireString(value *string, key string) (string, error) {
	if value == nil {
		return "", fieldMissing(key)
	}

	return *value, nil
}

func requireUint64(value *uint64, key string) (uint64, error) {
	if value == nil {
		return 0, fieldMissing(key)
	}

	return *value, nil
}

func requireInt64(value *int64, key string) (int64, error) {
	if value == nil {
		return 0, fieldMissing(key)
	}

	return *value, nil
}

func fieldMissing(key string) error {
	return fmt.Errorf("%w: missing %s", errCLIVersionInvalid, key)
}

func fieldInvalid(key string) error {
	return fmt.Errorf("%w: invalid %s", errCLIVersionInvalid, key)
}
