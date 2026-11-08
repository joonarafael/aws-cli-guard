//go:build unix

package install

import (
	"fmt"
	"path/filepath"
)

func installSiblingsAndRealAWS(
	env []string,
	siblingDir string,
	reserved []string,
) (string, string, string, error) {
	siblingAWS, err := regularSibling(filepath.Join(siblingDir, nameAWS))
	if err != nil {
		return "", "", "", err
	}

	siblingDisable, err := regularSibling(filepath.Join(siblingDir, nameDisableScript))
	if err != nil {
		return "", "", "", err
	}

	realAWS, err := DiscoverRealAWS(discoverPathEnv(env), siblingAWS)
	if err != nil {
		return "", "", "", err
	}

	conflict, err := realAWSConflictsPrefix(realAWS, reserved)
	if err != nil {
		return "", "", "", fmt.Errorf("aws-cli-guard: real AWS CLI path: %w", err)
	}

	if conflict {
		return "", "", "", ErrRealAWSInPrefix
	}

	return siblingAWS, siblingDisable, realAWS, nil
}
