// Package files provides safe on-disk operations for AWS CLI Guard: private directories,
// atomic file replacement, symlink refusal, and advisory locks for grants rewrites.
//
// Grants read-modify-write composition (WriteAtomic does not lock):
//
//	lock, err := files.OpenLock(grantsPath)
//	// check err before Release; Release is not nil-safe
//	defer lock.Release()
//	current, err := files.ReadRegular(grantsPath) // missing file means no grants
//	// modify current
//	err = files.WriteAtomic(grantsPath, next)
package files
