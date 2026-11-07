//go:build unix

package install

import (
	"aws-cli-guard/internal/version"
	"strings"
)

type manifestPaths struct {
	aws           string
	awsGuard      string
	realAWS       string
	disableScript string
	aside         string
	marker        string
}

func encodeManifest(paths manifestPaths) []byte {
	lines := []string{
		"version=" + version.Version,
		"aws=" + paths.aws,
		"aws_guard=" + paths.awsGuard,
		"real_aws=" + paths.realAWS,
		"disable_script=" + paths.disableScript,
		"aside=" + paths.aside,
		"marker=" + paths.marker,
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}
