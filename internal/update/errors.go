package update

import "errors"

var (
	errDownloadHTTP         = errors.New("aws-cli-guard: update: download failed")
	errDownloadTooLarge     = errors.New("aws-cli-guard: update: download too large")
	errChecksumsTooLarge    = errors.New("aws-cli-guard: update: checksums too large")
	errChecksumLineMissing  = errors.New("aws-cli-guard: update: checksum line missing")
	errChecksumLineInvalid  = errors.New("aws-cli-guard: update: checksum line invalid")
	errChecksumMismatch     = errors.New("aws-cli-guard: update: checksum mismatch")
	errArchiveMemberRegular = errors.New("aws-cli-guard: update: archive member not regular")
	errArchiveMemberCount   = errors.New("aws-cli-guard: update: archive member count")
	errArchiveMemberMissing = errors.New("aws-cli-guard: update: archive member missing")
	errArchiveAbsolutePath  = errors.New("aws-cli-guard: update: absolute path")
	errArchiveDirectory     = errors.New("aws-cli-guard: update: directory member")
	errArchiveParentPath    = errors.New("aws-cli-guard: update: parent path")
	errArchiveUnexpected    = errors.New("aws-cli-guard: update: unexpected member")
	errReleaseHTTP          = errors.New("aws-cli-guard: update: release request failed")
)
