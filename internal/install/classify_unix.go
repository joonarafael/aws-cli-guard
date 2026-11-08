//go:build unix

package install

import (
	"path/filepath"
)

type prefixKind int

const (
	prefixFresh prefixKind = iota
	prefixUpgrade
)

func classifyPrefix(prefix string) (prefixKind, error) {
	layout := layoutPaths(prefix)
	manifestPath := filepath.Join(prefix, nameManifest)

	core := []string{
		layout.aws,
		layout.awsGuard,
		layout.disableScript,
		manifestPath,
	}

	asideOcc, err := pathOccupied(layout.aside)
	if err != nil {
		return prefixFresh, err
	}

	if asideOcc {
		return prefixFresh, ErrDisableLayout
	}

	markerOcc, err := pathOccupied(layout.marker)
	if err != nil {
		return prefixFresh, err
	}

	if markerOcc {
		return prefixFresh, ErrDisableLayout
	}

	present := 0

	for _, path := range core {
		occupied, statErr := pathOccupied(path)
		if statErr != nil {
			return prefixFresh, statErr
		}

		if occupied {
			present++
		}
	}

	switch present {
	case 0:
		return prefixFresh, nil
	case len(core):
		return prefixUpgrade, nil
	default:
		return prefixFresh, ErrPathOccupied
	}
}
