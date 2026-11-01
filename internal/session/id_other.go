//go:build !unix

package session

import (
	"fmt"
	"os"
)

// ID returns a stable key for the current terminal session.
func ID() string {
	return fmt.Sprintf("ppid:%d", os.Getppid())
}
