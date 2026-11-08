package cli

import (
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"io"
	"path/filepath"
)

func writeUpgradeMigrateGuide(stderr io.Writer, env []string, programsReplaced bool) {
	if programsReplaced {
		output.Detail(
			stderr,
			"upgrade: install prefix programs were replaced; policy and grants were not deleted",
		)
	} else {
		output.Detail(stderr, "upgrade: nothing in the install prefix was changed")
	}

	output.Detail(
		stderr,
		"upgrade: if a rewrite started, the previous file is a sibling named *.migrated.*",
	)
	output.Detail(stderr, "upgrade: fix the file and run aws-guard migrate")
	output.Detail(stderr, "upgrade: for a corrupt file, run aws-guard recover policy")
	output.Detail(stderr, "upgrade: or aws-guard recover grants on a terminal")
	output.Detail(stderr, "upgrade: to start over, move config, state, and cache aside")
	output.Detail(stderr, "upgrade: then run aws-guard init")
	output.Detail(
		stderr,
		"upgrade: moving config also moves settings.toml; you can move that file back",
	)

	configPath, err := paths.ConfigFile(env)
	if err == nil {
		output.Detail(stderr, "upgrade: config directory: "+filepath.Dir(configPath))
	}

	grantsPath, err := paths.GrantsFile(env)
	if err == nil {
		output.Detail(stderr, "upgrade: state directory: "+filepath.Dir(grantsPath))
	}

	cacheDir, err := paths.CacheDir(env)
	if err == nil {
		output.Detail(stderr, "upgrade: cache directory: "+cacheDir)
	}

	output.Detail(
		stderr,
		"upgrade: for a broken install prefix, move that directory aside and run install again",
	)
}
