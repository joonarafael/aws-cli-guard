package grants

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"os"
	"path/filepath"
)

// RevokeRequest selects grants to remove. Set All to remove every record, or Profile to
// remove every record for that profile name.
type RevokeRequest struct {
	All     bool
	Profile string
}

// Revoke removes matching grant records under lock. It returns how many records were
// removed. When nothing was removed it does not create or rewrite the grants file.
func Revoke(env []string, req RevokeRequest) (int, error) {
	path, err := paths.GrantsFile(env)
	if err != nil {
		return 0, mapPathError(err)
	}

	parent := filepath.Dir(path)

	_, statErr := os.Lstat(parent)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return 0, nil
		}

		return 0, mapWriteError(path, statErr)
	}

	lock, err := files.OpenLock(path)
	if err != nil {
		return 0, mapWriteError(path, err)
	}

	current, loadErr := loadAtPath(path)
	if loadErr != nil {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return 0, mapReleaseError(path, releaseErr)
		}

		return 0, loadErr
	}

	next, removed := filterRevoke(current.Grants, req)
	if removed == 0 {
		releaseErr := lock.Release()
		if releaseErr != nil {
			return 0, mapReleaseError(path, releaseErr)
		}

		return 0, nil
	}

	writeErr := writeLocked(path, next)
	releaseErr := lock.Release()

	if writeErr != nil {
		return 0, writeErr
	}

	if releaseErr != nil {
		return 0, mapReleaseError(path, releaseErr)
	}

	return removed, nil
}

func filterRevoke(list []Grant, req RevokeRequest) ([]Grant, int) {
	if len(list) == 0 {
		return list, 0
	}

	if req.All {
		return []Grant{}, len(list)
	}

	removed := 0
	kept := make([]Grant, 0, len(list))

	for _, grant := range list {
		if grant.Profile == req.Profile {
			removed++

			continue
		}

		kept = append(kept, grant)
	}

	return kept, removed
}
