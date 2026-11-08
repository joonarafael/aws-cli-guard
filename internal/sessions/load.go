package sessions

import (
	"aws-cli-guard/internal/decode"
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"os"
)

const (
	maxFileSize  = 1 << 20
	writeVersion = 1
)

func supportedSessionVersions() []int {
	return []int{writeVersion}
}

// Load reads and validates the sessions file from env.
func Load(env []string) (File, error) {
	path, err := paths.SessionsFile(env)
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
				Version: writeVersion,
				Mode:    0,
				Records: []Record{},
			}, nil
		}

		return File{}, mapReadError(path, err)
	}

	if len(data) > maxFileSize {
		return File{}, newError(KindRead, path, "", 0, 0, ErrTooLarge)
	}

	var doc document

	decodeErr := decode.Decode(decode.JSON, data, &doc, supportedSessionVersions())
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

	return fmt.Errorf("aws-cli-guard: sessions: resolve sessions path: %w", err)
}

func mapReadError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindRead, path, "", 0, 0, err)
	}

	return newError(KindRead, path, "", 0, 0, err)
}

func mapDecodeError(path string, err error) error {
	if decErr, ok := errors.AsType[*decode.Error](err); ok {
		sessErr := newError(
			KindDecode,
			path,
			decErr.Key,
			decErr.Line,
			decErr.Column,
			decErr,
		)
		sessErr.Version = decErr.Version
		sessErr.HasVersion = decErr.HasVersion
		sessErr.Supported = append([]int(nil), decErr.Supported...)

		return sessErr
	}

	return newError(KindDecode, path, "", 0, 0, err)
}
