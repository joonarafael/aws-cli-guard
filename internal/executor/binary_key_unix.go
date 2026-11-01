//go:build unix

package executor

import (
	"aws-cli-guard/internal/cache"
	"syscall"
)

func cliVersionKey(realPath, sessionID string) (cache.CLIVersionKey, error) {
	var stat syscall.Stat_t

	err := syscall.Stat(realPath, &stat)
	if err != nil {
		return cache.CLIVersionKey{}, ErrStartFailed
	}

	return cache.CLIVersionKey{
		Session:   sessionID,
		Path:      realPath,
		Device:    stat.Dev,
		Inode:     stat.Ino,
		Size:      stat.Size,
		MtimeNsec: stat.Mtim.Nsec,
	}, nil
}
