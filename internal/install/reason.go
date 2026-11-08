package install

import "errors"

const (
	reasonHomeUnset           = "the home directory is unavailable"
	reasonPrefixInvalid       = "the install prefix is invalid"
	reasonPathOccupied        = "an install path is already occupied"
	reasonPrefixNotDirectory  = "the install prefix is not a directory"
	reasonSiblingMissing      = "a required install file is missing from this package"
	reasonRealAWSNotFound     = "the real AWS CLI was not found"
	reasonRealAWSInPrefix     = "the real AWS CLI path is inside the install prefix"
	reasonResolveExecutable   = "the running executable could not be resolved"
	reasonPlatformUnsupported = "install is not supported on this platform"
	reasonInstallFailed       = "install failed"
	reasonNotInstalled        = "there is no installation to upgrade in that prefix"
	reasonNotOurInstallation  = "the installed files are not a wrapper this binary can replace"
	reasonDisableLayout       = "persistent disable is active or partially present in that prefix"
	reasonManifestInvalid     = "the install manifest is invalid"
	reasonUpgradePartial      = "upgrade did not finish; run the same command again"
)

// Reason maps an install error to a fixed user-facing reason line.
func Reason(err error) string {
	switch {
	case errors.Is(err, ErrHomeUnset):
		return reasonHomeUnset
	case errors.Is(err, ErrPrefixInvalid):
		return reasonPrefixInvalid
	case errors.Is(err, ErrPathOccupied):
		return reasonPathOccupied
	case errors.Is(err, ErrPrefixNotDirectory):
		return reasonPrefixNotDirectory
	case errors.Is(err, ErrSiblingMissing):
		return reasonSiblingMissing
	case errors.Is(err, ErrRealAWSNotFound):
		return reasonRealAWSNotFound
	case errors.Is(err, ErrRealAWSInPrefix):
		return reasonRealAWSInPrefix
	case errors.Is(err, ErrResolveExecutable):
		return reasonResolveExecutable
	case errors.Is(err, ErrPlatformUnsupported):
		return reasonPlatformUnsupported
	case errors.Is(err, ErrNotInstalled):
		return reasonNotInstalled
	case errors.Is(err, ErrNotOurInstallation):
		return reasonNotOurInstallation
	case errors.Is(err, ErrDisableLayout):
		return reasonDisableLayout
	case errors.Is(err, ErrManifestInvalid):
		return reasonManifestInvalid
	case errors.Is(err, ErrUpgradePartial):
		return reasonUpgradePartial
	default:
		return reasonInstallFailed
	}
}
