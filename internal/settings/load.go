package settings

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
)

const (
	maxFileSize      = 1 << 20
	supportedVersion = 1
)

// Load reads and validates the settings file from env. A missing file returns defaults.
func Load(env []string) (File, error) {
	path, err := paths.SettingsFile(env)
	if err != nil {
		return File{}, mapPathError(err)
	}

	return loadAtPath(path)
}

func loadAtPath(path string) (File, error) {
	data, mode, err := files.ReadRegular(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{
				Path:    path,
				Version: supportedVersion,
				Mode:    0,
				Verify:  false,
				Level:   LevelTight,
			}, nil
		}

		return File{}, mapReadError(path, err)
	}

	if len(data) > maxFileSize {
		return File{}, newError(KindRead, path, "", 0, 0, ErrTooLarge)
	}

	var doc document

	decodeErr := decode.Decode(decode.TOML, data, &doc, []int{supportedVersion})
	if decodeErr != nil {
		return File{}, mapDecodeError(path, decodeErr)
	}

	out, valErr := validateSettings(path, doc)
	if valErr != nil {
		return File{}, stampInvalidError(valErr, path, doc.Version)
	}

	out.Mode = mode

	return out, nil
}

func mapPathError(err error) error {
	if errors.Is(err, paths.ErrHomeUnset) {
		return newError(KindPath, "", "", 0, 0, err)
	}

	return fmt.Errorf("aws-cli-guard: settings: resolve config path: %w", err)
}

func mapReadError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindRead, path, "", 0, 0, err)
	}

	return newError(KindRead, path, "", 0, 0, err)
}

func mapDecodeError(path string, err error) error {
	if decErr, ok := errors.AsType[*decode.Error](err); ok {
		setErr := newError(
			KindDecode,
			path,
			decErr.Key,
			decErr.Line,
			decErr.Column,
			decErr,
		)
		setErr.Version = decErr.Version
		setErr.HasVersion = decErr.HasVersion
		setErr.Supported = append([]int(nil), decErr.Supported...)

		return setErr
	}

	return newError(KindDecode, path, "", 0, 0, err)
}

func stampInvalidError(err error, path string, version int) error {
	setErr, ok := errors.AsType[*Error](err)
	if !ok {
		return err
	}

	if setErr.Path == "" {
		setErr.Path = stripControls(path)
	}

	setErr.Version = version
	setErr.HasVersion = true
	setErr.Supported = []int{supportedVersion}

	return setErr
}
