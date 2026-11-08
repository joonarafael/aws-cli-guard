//go:build linux

package session

import (
	"fmt"
	"os"
	"syscall"
)

func controllingTTYPath(fd int) (string, error) {
	link := fmt.Sprintf("/proc/self/fd/%d", fd)

	path, err := os.Readlink(link)
	if err != nil {
		return "", ErrNoTTY
	}

	return path, nil
}

func statDev(stat *syscall.Stat_t) uint64 {
	return stat.Dev
}

func statIno(stat *syscall.Stat_t) uint64 {
	return stat.Ino
}

func statRdev(stat *syscall.Stat_t) uint64 {
	return stat.Rdev
}
