package grants

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

// Update loads grants under lock, applies mutate, and writes the file.
func Update(env []string, mutate func([]Grant) ([]Grant, error)) error {
	path, err := paths.GrantsFile(env)
	if err != nil {
		return mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return mapWriteError(path, err)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return mapWriteError(path, err)
	}

	current, loadErr := loadAtPath(path)
	if loadErr != nil {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return mapReleaseError(path, releaseErr)
		}

		return loadErr
	}

	next, mutateErr := mutate(current.Grants)
	if mutateErr != nil {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return mapReleaseError(path, releaseErr)
		}

		return mutateErr
	}

	writeErr := writeLocked(path, next)
	releaseErr := lock.Release()

	if writeErr != nil {
		return writeErr
	}

	if releaseErr != nil {
		return mapReleaseError(path, releaseErr)
	}

	return nil
}

// Replace writes grants without reading the current file.
func Replace(env []string, list []Grant) error {
	path, err := paths.GrantsFile(env)
	if err != nil {
		return mapPathError(err)
	}

	err = files.MkdirPrivate(filepath.Dir(path))
	if err != nil {
		return mapWriteError(path, err)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return mapWriteError(path, err)
	}

	writeErr := writeLocked(path, list)
	releaseErr := lock.Release()

	if writeErr != nil {
		return writeErr
	}

	if releaseErr != nil {
		return mapReleaseError(path, releaseErr)
	}

	return nil
}

func writeLocked(path string, list []Grant) error {
	valErr := validateGrantList(path, list)
	if valErr != nil {
		return valErr
	}

	data, err := encodeGrants(list)
	if err != nil {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	err = files.WriteAtomic(path, data)
	if err != nil {
		return mapWriteError(path, err)
	}

	return nil
}

func encodeGrants(list []Grant) ([]byte, error) {
	if list == nil {
		list = []Grant{}
	}

	encoded := encodeDocument{
		Version: supportedVersion,
		Grants:  make([]encodeGrant, len(list)),
	}

	for i, grant := range list {
		encoded.Grants[i] = encodeGrant{
			Profile: grant.Profile,
			Expires: grant.Expires.UTC().Format(time.RFC3339),
			Source: encodeSource{
				AccountID:         grant.Source.AccountID,
				RoleARN:           grant.Source.RoleARN,
				SSOAccountID:      grant.Source.SSOAccountID,
				SSORoleName:       grant.Source.SSORoleName,
				CredentialProcess: grant.Source.CredentialProcess,
			},
		}
	}

	body, err := json.MarshalIndent(&encoded, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("aws-cli-guard: grants: encode json: %w", err)
	}

	body = append(body, '\n')

	return body, nil
}

func mapWriteError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: grants: %w", err))
}

func mapReleaseError(path string, err error) error {
	return newError(
		KindWrite,
		path,
		"",
		0,
		0,
		fmt.Errorf("aws-cli-guard: grants: release lock: %w", err),
	)
}
