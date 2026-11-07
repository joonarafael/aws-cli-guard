//go:build unix

package install

import (
	"io"
	"os"
	"path/filepath"
)

type installInputs struct {
	prefix          string
	layout          installLayout
	siblingAWS      string
	siblingDisable  string
	guardExecutable string
	realAWS         string
}

func writeInstallLayout(inputs installInputs, stdout io.Writer) error {
	manifestPath := filepath.Join(inputs.prefix, nameManifest)

	var created []string

	installOne := func(dst, src string, mode os.FileMode) error {
		err := copyFileExclusive(dst, src, mode)
		if err != nil {
			removeCreated(created)

			return err
		}

		created = append(created, dst)

		return nil
	}

	err := installOne(inputs.layout.aws, inputs.siblingAWS, modeExecutable)
	if err != nil {
		return err
	}

	err = installOne(inputs.layout.awsGuard, inputs.guardExecutable, modeExecutable)
	if err != nil {
		return err
	}

	err = installOne(inputs.layout.disableScript, inputs.siblingDisable, modeExecutable)
	if err != nil {
		return err
	}

	manifest := encodeManifest(manifestPaths{
		aws:           inputs.layout.aws,
		awsGuard:      inputs.layout.awsGuard,
		realAWS:       inputs.realAWS,
		disableScript: inputs.layout.disableScript,
		aside:         inputs.layout.aside,
		marker:        inputs.layout.marker,
	})

	err = writeBytesExclusive(manifestPath, manifest, modeManifest)
	if err != nil {
		removeCreated(created)

		return err
	}

	writeSuccess(stdout, inputs.prefix)

	return nil
}
