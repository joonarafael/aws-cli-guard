package cache_test

import (
	"aws-cli-guard/internal/cache"
	"aws-cli-guard/internal/cliversion"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const testCLIVersion209 = "2.9.0"

func TestLoadCLIVersionMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CACHE_DIR=" + dir}

	file, err := cache.LoadCLIVersion(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Entries) != 0 {
		t.Fatalf("entries = %v", file.Entries)
	}
}

func TestCLIVersionRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"AWS_CLI_GUARD_CACHE_DIR=" + dir}

	key := cache.CLIVersionKey{
		Session:   "sid:1",
		Path:      "/usr/bin/aws",
		Device:    8,
		Inode:     9,
		Size:      100,
		MtimeNsec: 200,
	}

	entry := cache.CLIVersionEntry{
		Key:        key,
		Relation:   cliversion.RelationSupported,
		CLIVersion: testCLIVersion209,
		Warned:     false,
	}

	err := cache.UpsertCLIVersion(env, entry)
	if err != nil {
		t.Fatal(err)
	}

	file, err := cache.LoadCLIVersion(env)
	if err != nil {
		t.Fatal(err)
	}

	got, found := cache.LookupCLIVersion(file, key)
	if !found {
		t.Fatal("expected hit")
	}

	if got.CLIVersion != testCLIVersion209 || got.Warned {
		t.Fatalf("got = %+v", got)
	}

	entry.Warned = true

	err = cache.UpsertCLIVersion(env, entry)
	if err != nil {
		t.Fatal(err)
	}

	file, err = cache.LoadCLIVersion(env)
	if err != nil {
		t.Fatal(err)
	}

	got, found = cache.LookupCLIVersion(file, key)
	if !found || !got.Warned {
		t.Fatalf("warned = %v", got.Warned)
	}
}

func TestLoadCLIVersionDuplicateKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cli-version.json")

	body := `{
  "version": 1,
  "entries": [
    {
      "session": "sid:1",
      "path": "/a",
      "device": 1,
      "inode": 2,
      "size": 3,
      "mtime_nsec": 4,
      "relation": "supported",
      "cli_version": "2.9.0",
      "warned": false
    },
    {
      "session": "sid:1",
      "path": "/a",
      "device": 1,
      "inode": 2,
      "size": 3,
      "mtime_nsec": 4,
      "relation": "supported",
      "cli_version": "2.9.0",
      "warned": true
    }
  ]
}
`

	err := os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.LoadCLIVersion([]string{"AWS_CLI_GUARD_CACHE_DIR=" + dir})
	if !errors.Is(err, cache.ErrCLIVersionUnusable) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadCLIVersionMissingWarned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cli-version.json")

	body := `{
  "version": 1,
  "entries": [
    {
      "session": "sid:1",
      "path": "/a",
      "device": 1,
      "inode": 2,
      "size": 3,
      "mtime_nsec": 4,
      "relation": "supported",
      "cli_version": "2.9.0"
    }
  ]
}
`

	err := os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.LoadCLIVersion([]string{"AWS_CLI_GUARD_CACHE_DIR=" + dir})
	if !errors.Is(err, cache.ErrCLIVersionUnusable) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpsertDoesNotRewriteUnusable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cli-version.json")

	before := []byte(`{not json`)

	err := os.WriteFile(path, before, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	env := []string{"AWS_CLI_GUARD_CACHE_DIR=" + dir}
	entry := cache.CLIVersionEntry{
		Key: cache.CLIVersionKey{
			Session: "sid:1", Path: "/a", Device: 1, Inode: 2, Size: 3, MtimeNsec: 4,
		},
		Relation:   cliversion.RelationSupported,
		CLIVersion: testCLIVersion209,
		Warned:     true,
	}

	err = cache.UpsertCLIVersion(env, entry)
	if !errors.Is(err, cache.ErrCLIVersionUnusable) {
		t.Fatalf("err = %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(after) != string(before) {
		t.Fatal("corrupt file was modified")
	}
}
