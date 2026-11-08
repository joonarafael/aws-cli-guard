//go:build unix

package install

import (
	"fmt"
	"os"
)

func verifyInstallationIdentity(record manifestRecord, siblingAWS, siblingGuard string) error {
	err := verifyWrapperFile(record.paths.aws, siblingAWS)
	if err != nil {
		return err
	}

	err = verifyWrapperFile(record.paths.awsGuard, siblingGuard)
	if err != nil {
		return err
	}

	return verifyDisableScript(record.paths.disableScript)
}

func verifyWrapperFile(installedPath, siblingPath string) error {
	err := refuseSymlink(installedPath)
	if err != nil {
		return ErrNotOurInstallation
	}

	installedPkg := packagePath(installedPath)

	siblingPkg := packagePath(siblingPath)
	if installedPkg == "" || siblingPkg == "" || installedPkg != siblingPkg {
		return ErrNotOurInstallation
	}

	return nil
}

func verifyDisableScript(path string) error {
	err := refuseSymlink(path)
	if err != nil {
		return ErrNotOurInstallation
	}

	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: stat disable script: %w", err)
	}

	if !info.Mode().IsRegular() {
		return ErrNotOurInstallation
	}

	return nil
}

func refuseSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: lstat %q: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return ErrNotOurInstallation
	}

	if !info.Mode().IsRegular() {
		return ErrNotOurInstallation
	}

	return nil
}
