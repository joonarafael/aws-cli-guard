package install

import "errors"

// ErrHomeUnset is returned when the default install prefix needs a home directory
// that cannot be resolved.
var ErrHomeUnset = errors.New("aws-cli-guard: home directory is unavailable")

// ErrPrefixInvalid is returned when the install prefix contains disallowed characters.
var ErrPrefixInvalid = errors.New("aws-cli-guard: install prefix is invalid")

// ErrPathOccupied is returned when a reserved install path already exists.
var ErrPathOccupied = errors.New("aws-cli-guard: install path is already occupied")

// ErrPrefixNotDirectory is returned when the install prefix exists and is not a directory.
var ErrPrefixNotDirectory = errors.New("aws-cli-guard: install prefix is not a directory")

// ErrSiblingMissing is returned when a required sibling of the running executable is absent.
var ErrSiblingMissing = errors.New("aws-cli-guard: install sibling is missing")

// ErrRealAWSNotFound is returned when no suitable real AWS CLI executable is found on PATH.
var ErrRealAWSNotFound = errors.New("aws-cli-guard: real AWS CLI not found in PATH")

// ErrRealAWSInPrefix is returned when the real AWS CLI path is inside the install prefix.
var ErrRealAWSInPrefix = errors.New("aws-cli-guard: real AWS CLI path is inside the install prefix")

// ErrResolveExecutable is returned when the running executable path cannot be resolved.
var ErrResolveExecutable = errors.New("aws-cli-guard: cannot resolve running executable")

// ErrPlatformUnsupported is returned on platforms without install support.
var ErrPlatformUnsupported = errors.New(
	"aws-cli-guard: install is not supported on this platform yet",
)

// ErrNotInstalled is returned when upgrade is run on a prefix with no installation.
var ErrNotInstalled = errors.New("aws-cli-guard: install prefix has no installation")

// ErrNotOurInstallation is returned when installed files are not wrappers this binary can replace.
var ErrNotOurInstallation = errors.New(
	"aws-cli-guard: installed files are not a wrapper this binary can replace",
)

// ErrDisableLayout is returned when persistent-disable paths are present in the prefix.
var ErrDisableLayout = errors.New(
	"aws-cli-guard: persistent-disable layout is present in the install prefix",
)

// ErrManifestInvalid is returned when the install manifest is missing or invalid.
var ErrManifestInvalid = errors.New("aws-cli-guard: install manifest is invalid")

// ErrUpgradePartial is returned when upgrade replaced some files but not all.
var ErrUpgradePartial = errors.New("aws-cli-guard: upgrade did not finish")

// ErrReplacePartial is returned when no partial filename is available for replace.
var ErrReplacePartial = errors.New("aws-cli-guard: replace: no partial name")
