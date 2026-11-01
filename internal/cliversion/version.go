package cliversion

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	relationOlderWire      = "older"
	relationSupportedWire  = "supported"
	relationNewerWire      = "newer"
	relationUnreadableWire = "unreadable"

	floorMajor = 2
	floorMinor = 9
	floorPatch = 0

	ceilingMajor = 2
	ceilingMinor = 9
	ceilingPatch = 0
)

// Relation is how a parsed CLI version relates to the supported range.
type Relation int

const (
	// RelationUnreadable means no comparable version was found.
	RelationUnreadable Relation = iota
	// RelationOlder is below the 2.9.0 floor.
	RelationOlder
	// RelationSupported is from the floor through the verified ceiling.
	RelationSupported
	// RelationNewer is above the verified ceiling.
	RelationNewer
)

// String returns the wire form stored in the version cache.
func (r Relation) String() string {
	switch r {
	case RelationOlder:
		return relationOlderWire
	case RelationSupported:
		return relationSupportedWire
	case RelationNewer:
		return relationNewerWire
	case RelationUnreadable:
		return relationUnreadableWire
	default:
		return relationUnreadableWire
	}
}

// ParseRelation parses a cache relation string.
func ParseRelation(raw string) (Relation, bool) {
	switch raw {
	case relationOlderWire:
		return RelationOlder, true
	case relationSupportedWire:
		return RelationSupported, true
	case relationNewerWire:
		return RelationNewer, true
	case relationUnreadableWire:
		return RelationUnreadable, true
	default:
		return RelationUnreadable, false
	}
}

// Result holds a parsed version and its relation to the supported range.
type Result struct {
	Relation Relation
	Version  string
}

// ClassifyOutput scans stdout then stderr for an aws-cli version token.
func ClassifyOutput(stdout, stderr []byte) Result {
	if ver, ok := parseFromOutput(stdout); ok {
		return classifyParsed(ver)
	}

	if ver, ok := parseFromOutput(stderr); ok {
		return classifyParsed(ver)
	}

	return Result{Relation: RelationUnreadable, Version: ""}
}

func classifyParsed(ver version) Result {
	rel := compareToRange(ver)

	return Result{
		Relation: rel,
		Version:  ver.normalized(),
	}
}

type version struct {
	parts []int
}

func (v version) normalized() string {
	if len(v.parts) == 0 {
		return ""
	}

	parts := make([]string, len(v.parts))
	for i, p := range v.parts {
		parts[i] = strconv.Itoa(p)
	}

	return strings.Join(parts, ".")
}

func parseFromOutput(data []byte) (version, bool) {
	text := string(data)

	for field := range strings.FieldsSeq(text) {
		ver, ok := parseToken(field)
		if ok {
			return ver, true
		}
	}

	return version{parts: nil}, false
}

func parseToken(token string) (version, bool) {
	const prefix = "aws-cli/"

	if !strings.HasPrefix(token, prefix) {
		return version{parts: nil}, false
	}

	rest := token[len(prefix):]
	if rest == "" {
		return version{parts: nil}, false
	}

	end := 0

	for end < len(rest) {
		char := rest[end]
		if char == '.' {
			end++

			continue
		}

		if char < '0' || char > '9' {
			break
		}

		end++
	}

	if end == 0 {
		return version{parts: nil}, false
	}

	segment := rest[:end]
	components := strings.Split(segment, ".")
	parts := make([]int, 0, len(components))

	for _, comp := range components {
		if comp == "" {
			return version{parts: nil}, false
		}

		n, err := strconv.Atoi(comp)
		if err != nil {
			return version{parts: nil}, false
		}

		parts = append(parts, n)
	}

	return version{parts: parts}, true
}

func compareToRange(ver version) Relation {
	if cmpVersion(ver, version{parts: []int{floorMajor, floorMinor, floorPatch}}) < 0 {
		return RelationOlder
	}

	if cmpVersion(ver, version{parts: []int{ceilingMajor, ceilingMinor, ceilingPatch}}) > 0 {
		return RelationNewer
	}

	return RelationSupported
}

func cmpVersion(left, right version) int {
	maxLen := max(len(right.parts), len(left.parts))

	for idx := range maxLen {
		leftVal := 0
		if idx < len(left.parts) {
			leftVal = left.parts[idx]
		}

		rightVal := 0
		if idx < len(right.parts) {
			rightVal = right.parts[idx]
		}

		if leftVal < rightVal {
			return -1
		}

		if leftVal > rightVal {
			return 1
		}
	}

	return 0
}

// PolicyApplies reports whether profile policy should run for this relation.
func PolicyApplies(rel Relation) bool {
	return rel == RelationSupported || rel == RelationNewer
}

// Warns reports whether this relation needs a version warning on stderr.
func Warns(rel Relation) bool {
	return rel == RelationOlder || rel == RelationNewer || rel == RelationUnreadable
}

// WarningText returns the secret-safe warning line for rel.
func WarningText(res Result) string {
	switch res.Relation {
	case RelationSupported:
		return ""
	case RelationOlder:
		if res.Version != "" {
			return fmt.Sprintf(
				"AWS CLI %s is older than 2.9.0; running it with no policy",
				res.Version,
			)
		}

		return "AWS CLI version could not be read; running it with no policy"
	case RelationUnreadable:
		return "AWS CLI version could not be read; running it with no policy"
	case RelationNewer:
		if res.Version != "" {
			return fmt.Sprintf(
				"AWS CLI %s is newer than the verified release 2.9.0; policy still applies",
				res.Version,
			)
		}

		return "AWS CLI is newer than the verified release 2.9.0; policy still applies"
	default:
		return ""
	}
}
