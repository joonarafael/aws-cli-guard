//go:build unix

package session

import (
	"os"
	"syscall"
)

func deviceNumbers(info os.FileInfo) (CharDeviceStat, bool) {
	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return CharDeviceStat{}, false //nolint:exhaustruct_v5 // zero value means absent.
	}

	return CharDeviceStat{
		Dev:  statDev(sys),
		Ino:  statIno(sys),
		Rdev: statRdev(sys),
	}, true
}
