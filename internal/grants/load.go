package grants

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

// Load reads and validates the grants file from env.
func Load(env []string) (File, error) {
	path, err := paths.GrantsFile(env)
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
				Grants:  []Grant{},
			}, nil
		}

		return File{}, mapReadError(path, err)
	}

	if len(data) > maxFileSize {
		return File{}, newError(KindRead, path, "", 0, 0, ErrTooLarge)
	}

	var doc document

	decodeErr := decode.Decode(decode.JSON, data, &doc, []int{supportedVersion})
	if decodeErr != nil {
		return File{}, mapDecodeError(path, decodeErr)
	}

	out, valErr := validateDocument(path, doc)
	if valErr != nil {
		return File{}, valErr
	}

	out.Mode = mode

	return out, nil
}

func mapPathError(err error) error {
	if errors.Is(err, paths.ErrHomeUnset) {
		return newError(KindPath, "", "", 0, 0, err)
	}

	return fmt.Errorf("aws-cli-guard: grants: resolve grants path: %w", err)
}

func mapReadError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindRead, path, "", 0, 0, err)
	}

	return newError(KindRead, path, "", 0, 0, err)
}

func mapDecodeError(path string, err error) error {
	if decErr, ok := errors.AsType[*decode.Error](err); ok {
		grErr := newError(
			KindDecode,
			path,
			decErr.Key,
			decErr.Line,
			decErr.Column,
			decErr,
		)
		grErr.Version = decErr.Version
		grErr.HasVersion = decErr.HasVersion
		grErr.Supported = append([]int(nil), decErr.Supported...)

		return grErr
	}

	return newError(KindDecode, path, "", 0, 0, err)
}

func stampInvalidError(err error, path string, version int) error {
	grErr, ok := errors.AsType[*Error](err)
	if !ok {
		return err
	}

	if grErr.Path == "" {
		grErr.Path = stripControls(path)
	}

	grErr.Version = version
	grErr.HasVersion = true
	grErr.Supported = []int{supportedVersion}

	return grErr
}
