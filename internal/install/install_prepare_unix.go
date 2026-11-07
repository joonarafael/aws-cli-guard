//go:build unix

package install

import (
	"fmt"
	"path/filepath"
)

func prepareInstall(opts Options) (installInputs, error) {
	executable, err := resolveExecutable(opts.Executable)
	if err != nil {
		return installInputs{}, err
	}

	prefix, err := resolvePrefix(opts.Env, opts.Prefix)
	if err != nil {
		return installInputs{}, err
	}

	err = validatePrefixChars(prefix)
	if err != nil {
		return installInputs{}, err
	}

	reserved := reservedInstallPaths(prefix)

	err = refuseOccupied(reserved)
	if err != nil {
		return installInputs{}, err
	}

	siblingDir := filepath.Dir(executable)

	siblingAWS, siblingDisable, realAWS, err := installSiblingsAndRealAWS(
		opts.Env,
		siblingDir,
		reserved,
	)
	if err != nil {
		return installInputs{}, err
	}

	err = ensurePrefixDirectory(prefix)
	if err != nil {
		return installInputs{}, err
	}

	return installInputs{
		prefix:          prefix,
		layout:          layoutPaths(prefix),
		siblingAWS:      siblingAWS,
		siblingDisable:  siblingDisable,
		guardExecutable: executable,
		realAWS:         realAWS,
	}, nil
}

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
