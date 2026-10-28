package policy

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

// Load reads and validates the user policy file from env.
func Load(env []string) (Policy, error) {
	path, err := paths.ConfigFile(env)
	if err != nil {
		return Policy{}, mapPathError(err)
	}

	data, mode, err := files.ReadRegular(path)
	if err != nil {
		return Policy{}, mapReadError(path, err)
	}

	if len(data) > maxFileSize {
		return Policy{}, newError(KindRead, path, "", 0, 0, ErrTooLarge)
	}

	var doc document

	decodeErr := decode.Decode(decode.TOML, data, &doc, []int{supportedVersion})
	if decodeErr != nil {
		return Policy{}, mapDecodeError(path, decodeErr)
	}

	out, valErr := validatePolicy(path, doc)
	if valErr != nil {
		return Policy{}, stampInvalidError(valErr, path, doc.Version)
	}

	out.Mode = mode

	return out, nil
}

func mapPathError(err error) error {
	if errors.Is(err, paths.ErrHomeUnset) {
		return newError(KindPath, "", "", 0, 0, err)
	}

	return fmt.Errorf("aws-cli-guard: policy: resolve config path: %w", err)
}

func mapReadError(path string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return newError(
			KindMissing,
			path,
			"",
			0,
			0,
			fmt.Errorf("%w: %w", ErrMissing, os.ErrNotExist),
		)
	}

	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindRead, path, "", 0, 0, err)
	}

	return newError(KindRead, path, "", 0, 0, err)
}

func mapDecodeError(path string, err error) error {
	if decErr, ok := errors.AsType[*decode.Error](err); ok {
		polErr := newError(
			KindDecode,
			path,
			decErr.Key,
			decErr.Line,
			decErr.Column,
			decErr,
		)
		polErr.Version = decErr.Version
		polErr.HasVersion = decErr.HasVersion
		polErr.Supported = append([]int(nil), decErr.Supported...)

		return polErr
	}

	return newError(KindDecode, path, "", 0, 0, err)
}

func stampInvalidError(err error, path string, version int) error {
	polErr, ok := errors.AsType[*Error](err)
	if !ok {
		return err
	}

	if polErr.Path == "" {
		polErr.Path = stripControls(path)
	}

	polErr.Version = version
	polErr.HasVersion = true
	polErr.Supported = []int{supportedVersion}

	return polErr
}
