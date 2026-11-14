//go:build !unix

package install

// ReplaceRelease is unsupported off unix.
func ReplaceRelease(_, _, _ string) error {
	return ErrPlatformUnsupported
}

// VerifiedAWSPrefix is unsupported off unix.
func VerifiedAWSPrefix(_ string) (string, bool) {
	return "", false
}

// PrefixUpgradeable is unsupported off unix.
func PrefixUpgradeable(_ string) error {
	return ErrPlatformUnsupported
}
