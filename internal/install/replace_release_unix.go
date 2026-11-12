//go:build unix

package install

import (
	"path/filepath"
)

// ReplaceRelease copies three programs from sourceDir into prefix and updates the manifest stamp.
//
//nolint:funlen // sequential replace steps match upgrade layout.
func ReplaceRelease(prefix, sourceDir, releaseVersion string) error {
	err := validatePrefixChars(prefix)
	if err != nil {
		return err
	}

	kind, err := classifyPrefix(prefix)
	if err != nil {
		return err
	}

	if kind != prefixUpgrade {
		return ErrNotInstalled
	}

	record, err := loadManifestRecord(prefix)
	if err != nil {
		return err
	}

	layout := layoutPaths(prefix)

	sourceAWS := filepath.Join(sourceDir, nameAWS)
	sourceGuard := filepath.Join(sourceDir, nameAWSGuard)
	sourceDisable := filepath.Join(sourceDir, nameDisableScript)

	err = verifyWrapperFile(layout.aws, sourceAWS)
	if err != nil {
		return err
	}

	err = verifyWrapperFile(layout.awsGuard, sourceGuard)
	if err != nil {
		return err
	}

	err = verifyDisableScript(layout.disableScript)
	if err != nil {
		return err
	}

	err = replaceRegularFile(layout.aws, sourceAWS, modeExecutable)
	if err != nil {
		return err
	}

	err = replaceRegularFile(layout.awsGuard, sourceGuard, modeExecutable)
	if err != nil {
		return ErrUpgradePartial
	}

	err = replaceRegularFile(layout.disableScript, sourceDisable, modeExecutable)
	if err != nil {
		return ErrUpgradePartial
	}

	manifest := encodeManifest(record.paths, releaseVersion)
	manifestPath := filepath.Join(prefix, nameManifest)

	err = replaceManifestFile(manifestPath, manifest)
	if err != nil {
		return ErrUpgradePartial
	}

	return nil
}

// PrefixUpgradeable reports whether prefix is a verified installation ready to replace.
func PrefixUpgradeable(prefix string) error {
	kind, err := classifyPrefix(prefix)
	if err != nil {
		return err
	}

	if kind != prefixUpgrade {
		return ErrNotInstalled
	}

	return nil
}

// VerifiedAWSPrefix reports whether executable is the manifest aws under a verified prefix.
func VerifiedAWSPrefix(executable string) (string, bool) {
	selfDir := filepath.Dir(executable)

	record, err := loadManifestRecord(selfDir)
	if err != nil {
		return "", false
	}

	classify, classErr := classifyPrefix(selfDir)
	if classErr != nil || classify != prefixUpgrade {
		return "", false
	}

	manifestAWS := record.paths.aws

	same, err := sameRegularFile(executable, manifestAWS)
	if err != nil || !same {
		return "", false
	}

	return selfDir, true
}

func sameRegularFile(left, right string) (bool, error) {
	err := refuseSymlink(left)
	if err != nil {
		return false, err
	}

	err = refuseSymlink(right)
	if err != nil {
		return false, err
	}

	return sameInstallPath(left, right), nil
}
