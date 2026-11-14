package gate

import (
	"aws-cli-guard/internal/update"
	"aws-cli-guard/internal/version"
	"io"
	"os"
	"time"
)

func runUpdateHandoff(stderr io.Writer, env []string) {
	stamp := version.Version
	if !update.ReleaseStamp(stamp) {
		return
	}

	executable, err := os.Executable()
	if err != nil {
		return
	}

	update.BeforeHandoff(stamp, executable, env, stderr, time.Now(), func() error {
		return update.StartDetachedCheck(executable, env)
	})
}
