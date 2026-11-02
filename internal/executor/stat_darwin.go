//go:build darwin

package executor

import "syscall"

func statDevice(stat *syscall.Stat_t) uint64 {
	//nolint:gosec // G115: darwin st_dev is int32; preserve device bits for cache keys.
	return uint64(uint32(stat.Dev))
}

func statMtimeNsec(stat *syscall.Stat_t) int64 {
	return stat.Mtimespec.Nano()
}
