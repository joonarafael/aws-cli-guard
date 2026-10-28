package policy_test

import (
	"aws-cli-guard/internal/policy"
	"os"
	"path/filepath"
	"testing"
)

func TestCorruptionSymlink(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	target := filepath.Join(dir, "real.toml")

	err = os.WriteFile(target, []byte("version = 1\n"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	link := filepath.Join(dir, "config.toml")

	err = os.Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	assertCorruptionFailure(t, env, "symlink", noCorruptionExpect())
}

func TestCorruptionDirectory(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	configPath := filepath.Join(dir, "config.toml")

	err = os.Mkdir(configPath, 0o700)
	if err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	assertCorruptionFailure(t, env, "not a regular file", noCorruptionExpect())
}

func TestCorruptionTooLarge(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)

	body := make([]byte, (1<<20)+1)
	for i := range body {
		body[i] = ' '
	}

	copy(body, []byte("version = 1\n"))

	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	path := filepath.Join(dir, "config.toml")

	err = os.WriteFile(path, body, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	assertCorruptionFailure(t, env, "file too large", noCorruptionExpect())
}

func TestCorruptionUnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root may read mode 000 files")
	}

	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\n")

	dir := policyDirFromEnv(t, env)
	path := filepath.Join(dir, "config.toml")

	err := os.Chmod(path, 0o000)
	if err != nil {
		t.Fatalf("Chmod: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(path, 0o600)
	})

	_, err = policy.Load(env)
	if err == nil {
		t.Skip("process could still read mode 000 file")
	}

	_, ok := policy.Corruption(err)
	if !ok {
		t.Fatalf("Corruption = false, want true for unreadable file: %v", err)
	}

	note, _ := policy.Corruption(err)
	if note.Failure != "unreadable" {
		t.Fatalf("failure = %q, want unreadable", note.Failure)
	}
}
