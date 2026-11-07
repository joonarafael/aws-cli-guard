//go:build unix

//nolint:testpackage // exercises relative prefix resolution against the process working directory.
package install

import (
	"path/filepath"
	"testing"
)

func TestResolvePrefixRelativeCwd(t *testing.T) { //nolint:paralleltest // uses t.Chdir
	work := t.TempDir()

	t.Chdir(work)

	got, err := resolvePrefix(nil, filepath.Join("nested", "prefix"))
	if err != nil {
		t.Fatalf("resolvePrefix: %v", err)
	}

	want, err := filepath.Abs(filepath.Join("nested", "prefix"))
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
