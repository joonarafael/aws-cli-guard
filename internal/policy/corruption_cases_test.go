package policy_test

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestCorruptionMissingFile(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	dir := policyDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, err = policy.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	_, ok := policy.Corruption(err)
	if ok {
		t.Fatal("Corruption = true, want false for missing file")
	}
}

func TestCorruptionHomeUnset(t *testing.T) {
	t.Parallel()

	_, err := policy.Load(nil)
	if err == nil {
		t.Fatal("expected error")
	}

	_, ok := policy.Corruption(err)
	if ok {
		t.Fatal("Corruption = true, want false for unset HOME")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("errors.Is ErrHomeUnset = false, got %v", err)
	}
}

func TestLoadStoresFileMode(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\n")

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Mode.Perm() != 0o600 {
		t.Fatalf("Mode = %o, want 0600", got.Mode.Perm())
	}
}

func TestLoosePermissionsWarnsAfterLoad(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicyPerm(t, env, "version = 1\n", 0o644)

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Mode.Perm() != 0o644 {
		t.Fatalf("Mode = %o, want 0644", got.Mode.Perm())
	}

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.ClassPolicy(), got.Path, got.Mode)

	if buf.Len() == 0 {
		t.Fatal("expected loose-permission warning")
	}
}

func TestCorruptionEmptyFile(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "")

	assertCorruptionFailure(t, env, "invalid syntax", noCorruptionExpect())
}

func TestCorruptionDuplicateKey(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 1\nversion = 2\n")

	assertCorruptionFailure(t, env, "duplicate key", noCorruptionExpect())
}

func TestCorruptionMissingVersion(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `[defaults]
action = "prompt"
`)

	assertCorruptionFailure(t, env, "missing version", corruptionExpect{
		key:        "version",
		path:       "",
		hasVersion: false,
		version:    0,
		supported:  "",
		omitSecret: false,
	})
}

func TestCorruptionVersionNotInteger(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = "one"
`)

	assertCorruptionFailure(t, env, "version is not an integer", corruptionExpect{
		key:        "version",
		path:       "",
		hasVersion: false,
		version:    0,
		supported:  "",
		omitSecret: false,
	})
}

func TestCorruptionUnsupportedVersion(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "version = 99\n")

	assertCorruptionFailure(t, env, "unsupported version", corruptionExpect{
		key:        "",
		path:       "",
		hasVersion: true,
		version:    99,
		supported:  "1",
		omitSecret: false,
	})
}

func TestCorruptionUnknownField(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[agents]
default = "deny"
`)

	assertCorruptionFailure(t, env, "unknown field", noCorruptionExpect())
}

func TestCorruptionInvalidValueType(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = 42
`)

	assertCorruptionFailure(t, env, "invalid value type", noCorruptionExpect())
}

func TestCorruptionInvalidAction(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[profiles.x]
action = "`+fixtureSecret+`"
`)

	path := configPath(t, env)

	assertCorruptionFailure(t, env, "invalid value", corruptionExpect{
		key:        "profiles.x.action",
		path:       path,
		hasVersion: true,
		version:    1,
		supported:  "1",
		omitSecret: true,
	})
}

func TestCorruptionSyntaxOmitsSecret(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, "not-valid = [[[\n"+fixtureSecret+"\n")

	assertCorruptionFailure(t, env, "invalid syntax", corruptionExpect{
		key:        "",
		path:       "",
		hasVersion: false,
		version:    0,
		supported:  "",
		omitSecret: true,
	})
}
