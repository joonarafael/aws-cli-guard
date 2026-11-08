package sessions

import (
	"aws-cli-guard/internal/paths"
	"fmt"
	"os"
)

// Present reports whether the sessions file path exists. It does not read or decode the file.
func Present(env []string) error {
	path, err := paths.SessionsFile(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: sessions: resolve sessions path: %w", err)
	}

	_, err = os.Lstat(path)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		return newError(
			KindMissing,
			path,
			"",
			0,
			0,
			fmt.Errorf("%w: %w", ErrMissing, os.ErrNotExist),
		)
	}

	return fmt.Errorf("aws-cli-guard: sessions: stat: %w", err)
}
