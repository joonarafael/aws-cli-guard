package policy

import (
	"aws-cli-guard/internal/paths"
	"fmt"
	"os"
)

// Present reports whether the policy file path exists. It does not read or decode the file.
// Nil means some directory entry exists. ErrMissing means the file is absent.
func Present(env []string) error {
	path, err := paths.ConfigFile(env)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: policy: resolve config path: %w", err)
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

	return fmt.Errorf("aws-cli-guard: policy: stat: %w", err)
}
