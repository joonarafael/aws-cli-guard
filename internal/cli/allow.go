package cli

import (
	"aws-cli-guard/internal/grants"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
)

const (
	allowCommand = "allow"
	allowMinArgs = 3

	reasonAllowUsage        = "usage: aws-guard allow <profile> [--ttl <duration>]"
	reasonProfileInvalid    = "the profile name is invalid"
	reasonTTLInvalid        = "the ttl is invalid"
	reasonTTLRequired       = "a ttl is required"
	reasonPolicyRead        = "the policy file could not be read"
	reasonAWSConfigRead     = "the aws config could not be read"
	reasonGrantNotConfirmed = "the grant was not confirmed"
	reasonAllowPath         = "the file path could not be resolved"
	reasonGrantsRead        = "the grants file could not be read"
	reasonGrantsWrite       = "the grants file could not be written"

	allowConfirmPrompt = "type the profile name to allow it on this installation until it expires"
	allowHeadlessNote  = "the grant applies to later invocations on this installation, " +
		"including headless ones, until it expires"
)

type allowArgs struct {
	profile string
	ttl     string
	hasTTL  bool
}

func runAllow(args []string, env []string, stdout, stderr io.Writer, confirm confirmFunc) int {
	parsed, ok := parseAllowArgs(args)
	if !ok {
		output.Error(stderr, reasonAllowUsage)

		return output.CancelStatus
	}

	if !validAllowProfile(parsed.profile) {
		output.Error(stderr, reasonProfileInvalid)

		return output.CancelStatus
	}

	ttl, ttlReason := resolveAllowTTL(env, parsed.profile, parsed)
	if ttlReason != "" {
		output.Error(stderr, ttlReason)

		return output.CancelStatus
	}

	source, err := grants.Fingerprint(env, parsed.profile)
	if err != nil {
		output.Error(stderr, reasonAWSConfigRead)

		return output.CancelStatus
	}

	err = confirm(parsed.profile, allowConfirmPrompt)
	if err != nil {
		output.Error(stderr, reasonGrantNotConfirmed)

		return output.CancelStatus
	}

	expires := time.Now().UTC().Add(ttl).Truncate(time.Second)

	err = grants.Update(env, func(list []grants.Grant) ([]grants.Grant, error) {
		grant := grants.Grant{
			Profile: parsed.profile,
			Expires: expires,
			Source:  source,
		}

		return upsertGrant(list, grant), nil
	})
	if err != nil {
		output.Error(stderr, allowWriteReason(err))

		return output.CancelStatus
	}

	writeAllowSuccess(stdout, parsed.profile, expires)

	return 0
}

func parseAllowArgs(args []string) (allowArgs, bool) {
	empty := allowArgs{profile: "", ttl: "", hasTTL: false}

	if len(args) < allowMinArgs {
		return empty, false
	}

	profile := args[2]
	if profile == "" || strings.HasPrefix(profile, "-") {
		return empty, false
	}

	rest := args[3:]
	if len(rest) == 0 {
		return allowArgs{profile: profile, ttl: "", hasTTL: false}, true
	}

	if len(rest) == 1 {
		ttl, ok := parseTTLFlag(rest[0])
		if !ok {
			return empty, false
		}

		return allowArgs{profile: profile, ttl: ttl, hasTTL: true}, true
	}

	if len(rest) == 2 && rest[0] == "--ttl" {
		if rest[1] == "" || strings.HasPrefix(rest[1], "-") {
			return empty, false
		}

		return allowArgs{profile: profile, ttl: rest[1], hasTTL: true}, true
	}

	return empty, false
}

func parseTTLFlag(token string) (string, bool) {
	const prefix = "--ttl="

	if after, ok := strings.CutPrefix(token, prefix); ok {
		if after == "" {
			return "", false
		}

		return after, true
	}

	return "", false
}

func validAllowProfile(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}

	return true
}

func resolveAllowTTL(env []string, profile string, parsed allowArgs) (time.Duration, string) {
	if parsed.hasTTL {
		dur, ok := parseAllowDuration(parsed.ttl)
		if !ok {
			return 0, reasonTTLInvalid
		}

		return dur, ""
	}

	pol, err := policy.Load(env)
	if err != nil {
		if errors.Is(err, policy.ErrMissing) {
			return 0, reasonTTLRequired
		}

		return 0, reasonPolicyRead
	}

	prof, ok := pol.Profiles[profile]
	if !ok || !prof.HasApprovalTTL {
		return 0, reasonTTLRequired
	}

	return prof.ApprovalTTL, ""
}

func parseAllowDuration(raw string) (time.Duration, bool) {
	dur, err := time.ParseDuration(raw)
	if err != nil || dur <= 0 {
		return 0, false
	}

	return dur, true
}

func upsertGrant(list []grants.Grant, grant grants.Grant) []grants.Grant {
	out := append([]grants.Grant{}, list...)

	for i, item := range out {
		if item.Profile == grant.Profile {
			out[i] = grant

			return out
		}
	}

	return append(out, grant)
}

func writeAllowSuccess(stdout io.Writer, profile string, expires time.Time) {
	output.Detail(stdout, "profile: "+profile)
	output.Detail(stdout, "expires: "+expires.Format(time.RFC3339))
	output.Detail(stdout, allowHeadlessNote)
}

func allowWriteReason(err error) string {
	if errors.Is(err, paths.ErrHomeUnset) {
		return reasonAllowPath
	}

	if grantErr, ok := errors.AsType[*grants.Error](err); ok {
		switch grantErr.Kind {
		case grants.KindMissing:
			return reasonGrantsRead
		case grants.KindPath:
			return reasonAllowPath
		case grants.KindRead, grants.KindDecode, grants.KindInvalid:
			return reasonGrantsRead
		case grants.KindWrite:
			return reasonGrantsWrite
		}
	}

	return reasonGrantsWrite
}
