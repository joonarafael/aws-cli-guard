package settings

import (
	"aws-cli-guard/internal/files"
	"aws-cli-guard/internal/paths"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultFileBodyWithUpdatesOff = `version = 1

[identity]
verify = false

[updates]
auto = false
`

const defaultFileBodyWithUpdatesOn = `version = 1

[identity]
verify = false

[updates]
auto = true
`

// EncodeBody builds settings.toml bytes from loaded fields plus the chosen auto value.
func EncodeBody(version int, verify bool, level Level, auto AutoChoice) ([]byte, error) {
	if auto == AutoAbsent {
		return nil, errEncodeAutoAbsent
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "version = %d\n", version)

	buf.WriteString("\n[identity]\n")

	if verify {
		fmt.Fprintln(&buf, "verify = true")
	} else {
		fmt.Fprintln(&buf, "verify = false")
	}

	if level != LevelTight {
		buf.WriteString("\n[logging]\n")
		fmt.Fprintf(&buf, "level = %q\n", level)
	}

	buf.WriteString("\n[updates]\n")

	if auto == AutoOn {
		fmt.Fprintln(&buf, "auto = true")
	} else {
		fmt.Fprintln(&buf, "auto = false")
	}

	return buf.Bytes(), nil
}

// DefaultBodyWithAuto is the init default plus [updates] for a new answer.
func DefaultBodyWithAuto(auto AutoChoice) string {
	if auto == AutoOn {
		return defaultFileBodyWithUpdatesOn
	}

	return defaultFileBodyWithUpdatesOff
}

// WriteAuto rewrites settings.toml preserving version, identity, and logging.
// It returns created true when the file did not exist before the write.
func WriteAuto(env []string, auto AutoChoice) (bool, error) {
	if auto == AutoAbsent {
		return false, errWriteAutoAbsent
	}

	path, err := paths.SettingsFile(env)
	if err != nil {
		return false, mapPathError(err)
	}

	_, statErr := os.Lstat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			err = writeSettingsAt(env, path, []byte(DefaultBodyWithAuto(auto)))
			if err != nil {
				return true, err
			}

			return true, nil
		}

		return false, mapWriteError(path, statErr)
	}

	current, err := loadAtPath(path)
	if err != nil {
		return false, err
	}

	body, err := EncodeBody(current.Version, current.Verify, current.Level, auto)
	if err != nil {
		return false, err
	}

	created := current.Auto == AutoAbsent

	err = writeSettingsAt(env, path, body)
	if err != nil {
		return created, err
	}

	return created, nil
}

func writeSettingsAt(env []string, path string, body []byte) error {
	dir := filepath.Dir(path)

	err := files.MkdirPrivate(dir)
	if err != nil {
		return mapWriteError(path, err)
	}

	err = files.WriteAtomic(path, body)
	if err != nil {
		return mapWriteError(path, err)
	}

	_ = env

	return nil
}

func mapWriteError(path string, err error) error {
	if errors.Is(err, files.ErrSymlink) || errors.Is(err, files.ErrNotRegular) {
		return newError(KindWrite, path, "", 0, 0, err)
	}

	return newError(KindWrite, path, "", 0, 0, fmt.Errorf("aws-cli-guard: settings: %w", err))
}
