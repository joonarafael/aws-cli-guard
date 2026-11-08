//go:build unix

package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func manifestKeys() []string {
	return []string{
		"version",
		"aws",
		"aws_guard",
		"real_aws",
		"disable_script",
		"aside",
		"marker",
	}
}

type manifestRecord struct {
	paths manifestPaths
}

func loadManifestRecord(prefix string) (manifestRecord, error) {
	manifestPath := filepath.Join(prefix, nameManifest)

	//nolint:gosec // G304: manifest path is under a verified install prefix.
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestRecord{}, ErrManifestInvalid
		}

		return manifestRecord{}, fmt.Errorf("aws-cli-guard: read manifest: %w", err)
	}

	return parseManifestContent(data, prefix)
}

func parseManifestContent(data []byte, prefix string) (manifestRecord, error) {
	text := strings.TrimSuffix(string(data), "\n")
	if text == "" {
		return manifestRecord{}, ErrManifestInvalid
	}

	lines := strings.Split(text, "\n")

	keys := manifestKeys()
	if len(lines) != len(keys) {
		return manifestRecord{}, ErrManifestInvalid
	}

	values := make(map[string]string, len(keys))

	for i, key := range keys {
		line := lines[i]

		gotKey, gotValue, ok := strings.Cut(line, "=")
		if !ok || gotKey != key || gotValue == "" {
			return manifestRecord{}, ErrManifestInvalid
		}

		values[key] = gotValue
	}

	layout := layoutPaths(prefix)
	want := manifestPaths{
		aws:           layout.aws,
		awsGuard:      layout.awsGuard,
		realAWS:       values["real_aws"],
		disableScript: layout.disableScript,
		aside:         layout.aside,
		marker:        layout.marker,
	}

	if values["aws"] != want.aws ||
		values["aws_guard"] != want.awsGuard ||
		values["disable_script"] != want.disableScript ||
		values["aside"] != want.aside ||
		values["marker"] != want.marker {
		return manifestRecord{}, ErrManifestInvalid
	}

	return manifestRecord{paths: want}, nil
}
