package cache

import "errors"

// ErrAbsent is returned when the cache directory does not exist.
var ErrAbsent = errors.New("aws-cli-guard: cache: directory absent")
