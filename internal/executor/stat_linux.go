//go:build linux

package executor

import "syscall"

func statDevice(stat *syscall.Stat_t) uint64 {
	return stat.Dev
}

func statMtimeNsec(stat *syscall.Stat_t) int64 {
	return stat.Mtim.Nano()
}
