//go:build !unix

package session

import "os"

func deviceNumbers(_ os.FileInfo) (CharDeviceStat, bool) {
	return CharDeviceStat{}, false //nolint:exhaustruct_v5 // zero value means absent.
}
