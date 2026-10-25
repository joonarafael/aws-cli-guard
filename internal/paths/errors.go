package paths

import "errors"

// ErrHomeUnset is returned when HOME and the relevant XDG variable are both unset
// and no override applies.
var ErrHomeUnset = errors.New(
	"aws-cli-guard: HOME is unset and no XDG or override directory is set",
)
