//go:build unix

package install

import (
	"io"
	"path/filepath"
)

type upgradeInputs struct {
	prefix          string
	layout          installLayout
	record          manifestRecord
	siblingAWS      string
	siblingDisable  string
	guardExecutable string
}

func runUpgrade(opts Options, stdout io.Writer) (Result, error) {
	executable, err := resolveExecutable(opts.Executable)
	if err != nil {
		return Result{}, err
	}

	prefix, err := resolvePrefix(opts.Env, opts.Prefix)
	if err != nil {
		return Result{}, err
	}

	err = validatePrefixChars(prefix)
	if err != nil {
		return Result{}, err
	}

	kind, err := classifyPrefix(prefix)
	if err != nil {
		return Result{}, err
	}

	if kind == prefixFresh {
		return Result{}, ErrNotInstalled
	}

	return executeUpgrade(opts, executable, prefix, stdout)
}

func executeUpgrade(opts Options, executable, prefix string, stdout io.Writer) (Result, error) {
	inputs, err := prepareUpgrade(executable, prefix)
	if err != nil {
		return Result{}, err
	}

	if opts.BeforeUpgrade != nil {
		err = opts.BeforeUpgrade()
		if err != nil {
			return Result{}, err
		}
	}

	err = writeUpgradeLayout(inputs, stdout)
	if err != nil {
		return Result{}, err
	}

	return Result{Upgraded: true, Prefix: prefix}, nil
}

func prepareUpgrade(executable, prefix string) (upgradeInputs, error) {
	record, err := loadManifestRecord(prefix)
	if err != nil {
		return upgradeInputs{}, err
	}

	siblingDir := filepath.Dir(executable)

	siblingAWS, err := regularSibling(filepath.Join(siblingDir, nameAWS))
	if err != nil {
		return upgradeInputs{}, err
	}

	siblingDisable, err := regularSibling(filepath.Join(siblingDir, nameDisableScript))
	if err != nil {
		return upgradeInputs{}, err
	}

	err = verifyInstallationIdentity(record, siblingAWS, executable)
	if err != nil {
		return upgradeInputs{}, err
	}

	err = ensurePrefixDirectory(prefix)
	if err != nil {
		return upgradeInputs{}, err
	}

	return upgradeInputs{
		prefix:          prefix,
		layout:          layoutPaths(prefix),
		record:          record,
		siblingAWS:      siblingAWS,
		siblingDisable:  siblingDisable,
		guardExecutable: executable,
	}, nil
}

func writeUpgradeLayout(inputs upgradeInputs, stdout io.Writer) error {
	err := replaceRegularFile(inputs.layout.aws, inputs.siblingAWS, modeExecutable)
	if err != nil {
		return err
	}

	err = replaceRegularFile(inputs.layout.awsGuard, inputs.guardExecutable, modeExecutable)
	if err != nil {
		return ErrUpgradePartial
	}

	err = replaceRegularFile(inputs.layout.disableScript, inputs.siblingDisable, modeExecutable)
	if err != nil {
		return ErrUpgradePartial
	}

	manifest := encodeManifest(inputs.record.paths)
	manifestPath := filepath.Join(inputs.prefix, nameManifest)

	err = replaceManifestFile(manifestPath, manifest)
	if err != nil {
		return ErrUpgradePartial
	}

	writeSuccess(stdout, inputs.prefix)

	return nil
}

func installOrUpgrade(opts Options, stdout io.Writer) (Result, error) {
	executable, err := resolveExecutable(opts.Executable)
	if err != nil {
		return Result{}, err
	}

	prefix, err := resolvePrefix(opts.Env, opts.Prefix)
	if err != nil {
		return Result{}, err
	}

	err = validatePrefixChars(prefix)
	if err != nil {
		return Result{}, err
	}

	kind, err := classifyPrefix(prefix)
	if err != nil {
		return Result{}, err
	}

	if kind == prefixUpgrade {
		return executeUpgrade(opts, executable, prefix, stdout)
	}

	inputs, err := prepareFreshInstall(opts, executable, prefix)
	if err != nil {
		return Result{}, err
	}

	err = writeInstallLayout(inputs, stdout)
	if err != nil {
		return Result{}, err
	}

	return Result{Upgraded: false, Prefix: prefix}, nil
}

func prepareFreshInstall(opts Options, executable, prefix string) (installInputs, error) {
	reserved := reservedInstallPaths(prefix)

	err := refuseOccupied(reserved)
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
