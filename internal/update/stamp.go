package update

import "regexp"

var releaseStampPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// ReleaseStamp reports whether stamp is a release version.
func ReleaseStamp(stamp string) bool {
	return releaseStampPattern.MatchString(stamp)
}
