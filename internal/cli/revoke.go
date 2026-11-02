package cli

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"errors"
	"io"
	"strconv"
)

const (
	revokeCommand = "revoke"
	revokeMinArgs = 3

	reasonRevokeUsage    = "usage: aws-guard revoke <profile>|--all"
	reasonRevokePath     = "the file path could not be resolved"
	reasonRevokeUnusable = "the grants file is unusable"
	reasonRevokeRead     = "the grants file could not be read"
	reasonRevokeWrite    = "the grants file could not be written"
)

type revokeArgs struct {
	all     bool
	profile string
}

func runRevoke(args []string, env []string, stdout, stderr io.Writer) int {
	parsed, ok := parseRevokeArgs(args)
	if !ok {
		output.Error(stderr, reasonRevokeUsage)

		return output.CancelStatus
	}

	if !parsed.all {
		if !validProfileName(parsed.profile) {
			output.Error(stderr, reasonProfileInvalid)

			return output.CancelStatus
		}
	}

	req := grants.RevokeRequest{All: parsed.all, Profile: parsed.profile}

	removed, err := grants.Revoke(env, req)
	if err != nil {
		return cancelRevokeGrants(stderr, err)
	}

	writeRevokeSuccess(stdout, parsed, removed)

	return 0
}

func parseRevokeArgs(args []string) (revokeArgs, bool) {
	empty := revokeArgs{all: false, profile: ""}

	if len(args) != revokeMinArgs {
		return empty, false
	}

	target := args[2]
	if target == "--all" {
		return revokeArgs{all: true, profile: ""}, true
	}

	if target == "" || target[0] == '-' {
		return empty, false
	}

	return revokeArgs{all: false, profile: target}, true
}

func writeRevokeSuccess(stdout io.Writer, parsed revokeArgs, removed int) {
	if parsed.all {
		if removed == 0 {
			output.Detail(stdout, "no grants")

			return
		}

		output.Detail(stdout, "revoked grants: "+strconv.Itoa(removed))

		return
	}

	if removed == 0 {
		output.Detail(stdout, "no grant for profile: "+parsed.profile)

		return
	}

	output.Detail(stdout, "revoked profile: "+parsed.profile)
}

func cancelRevokeGrants(stderr io.Writer, err error) int {
	if note, ok := grants.Corruption(err); ok {
		output.Error(stderr, reasonRevokeUnusable)
		diagnose.Write(stderr, note)

		return output.CancelStatus
	}

	if errors.Is(err, paths.ErrHomeUnset) {
		output.Error(stderr, reasonRevokePath)

		return output.CancelStatus
	}

	if grantErr, ok := errors.AsType[*grants.Error](err); ok {
		switch grantErr.Kind {
		case grants.KindPath:
			output.Error(stderr, reasonRevokePath)

			return output.CancelStatus
		case grants.KindWrite:
			output.Error(stderr, reasonRevokeWrite)

			return output.CancelStatus
		case grants.KindMissing, grants.KindRead, grants.KindDecode, grants.KindInvalid:
			output.Error(stderr, reasonRevokeRead)

			return output.CancelStatus
		}
	}

	output.Error(stderr, reasonRevokeRead)

	return output.CancelStatus
}
