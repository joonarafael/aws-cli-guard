package paths

import "errors"

// ErrHomeUnset is returned when the user-database lookup fails and HOME is unset,
// empty, or not absolute.
var ErrHomeUnset = errors.New("aws-cli-guard: home directory is unavailable")
