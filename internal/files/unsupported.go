//go:build !unix

package files

import "os"

// Lock is a placeholder on unsupported platforms.
type Lock struct{}

// MkdirPrivate creates a private directory tree.
func MkdirPrivate(_ string) error {
	return ErrPlatformUnsupported
}

// WriteAtomic atomically replaces a file.
func WriteAtomic(_ string, _ []byte) error {
	return ErrPlatformUnsupported
}

// ReadRegular reads a regular file.
func ReadRegular(_ string) ([]byte, os.FileMode, error) {
	return nil, 0, ErrPlatformUnsupported
}

// OpenLock acquires an advisory lock.
func OpenLock(_ string) (*Lock, error) {
	return nil, ErrPlatformUnsupported
}

// Release releases the lock.
func (lock *Lock) Release() error {
	_ = lock

	return ErrPlatformUnsupported
}
