package cli

import (
	"aws-cli-guard/internal/install"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"errors"
	"io"
)

const (
	shellCommand     = "shell"
	reasonShellUsage = "usage: aws-guard shell bash|zsh [--prefix dir] [--remove]"
	shellArgsMin     = 3
	shellNameBash    = "bash"
	shellNameZsh     = "zsh"
)

func runShell(args []string, env []string, stdout, stderr io.Writer) int {
	parsed, ok := parseShellArgs(args)
	if !ok {
		output.Error(stderr, reasonShellUsage)

		return output.CancelStatus
	}

	home, err := paths.HomeDir(env)
	if err != nil {
		output.Error(stderr, shellHomeReason(err))

		return output.CancelStatus
	}

	shellName, err := parseShellName(parsed.shell)
	if err != nil {
		output.Error(stderr, install.Reason(err))

		return output.CancelStatus
	}

	_, err = install.ApplyShell(install.ShellOptions{
		Env:    env,
		Home:   home,
		Prefix: parsed.prefix,
		Shell:  shellName,
		Remove: parsed.remove,
	}, stdout)
	if err != nil {
		output.Error(stderr, install.Reason(err))

		return output.CancelStatus
	}

	return 0
}

type shellArgs struct {
	shell  string
	prefix string
	remove bool
}

func parseShellArgs(args []string) (shellArgs, bool) {
	if len(args) < shellArgsMin {
		return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
	}

	if args[1] != shellCommand {
		return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
	}

	name := args[2]
	if name != shellNameBash && name != shellNameZsh {
		return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
	}

	rest := args[3:]

	var parsed shellArgs

	parsed.shell = name

	var index int

	if len(rest) > index && rest[index] == flagPrefix {
		if len(rest) < index+2 || rest[index+1] == "" {
			return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
		}

		parsed.prefix = rest[index+1]

		index += 2
	}

	if len(rest) > index {
		if rest[index] != flagRemove || len(rest) != index+1 {
			return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
		}

		parsed.remove = true
	}

	if len(rest) != index {
		return shellArgs{}, false //nolint:exhaustruct_v5 // invalid parse.
	}

	return parsed, true
}

const flagRemove = "--remove"

func parseShellName(name string) (install.ShellName, error) {
	switch name {
	case shellNameBash:
		return install.ShellBash, nil
	case shellNameZsh:
		return install.ShellZsh, nil
	default:
		return "", install.ErrShellNameUnsupported
	}
}

func shellHomeReason(err error) string {
	if errors.Is(err, paths.ErrHomeUnset) {
		return install.Reason(install.ErrHomeUnset)
	}

	return install.Reason(err)
}
