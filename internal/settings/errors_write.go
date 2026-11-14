package settings

import "errors"

var (
	errEncodeAutoAbsent = errors.New("aws-cli-guard: settings: encode: auto answer absent")
	errWriteAutoAbsent  = errors.New("aws-cli-guard: settings: write auto: absent answer")
)
