package policy_test

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/policy"
	"bytes"
	"strings"
	"testing"
)

type corruptionExpect struct {
	key        string
	path       string
	hasVersion bool
	version    int
	supported  string
	omitSecret bool
}

func noCorruptionExpect() corruptionExpect {
	return corruptionExpect{
		key:        "",
		path:       "",
		hasVersion: false,
		version:    0,
		supported:  "",
		omitSecret: false,
	}
}

func assertCorruptionFailure(
	t *testing.T,
	env []string,
	wantFailure string,
	expect corruptionExpect,
) {
	t.Helper()

	_, err := policy.Load(env)
	if err == nil {
		t.Fatal("expected load error")
	}

	note, ok := policy.Corruption(err)
	if !ok {
		t.Fatalf("Corruption = false, want true: %v", err)
	}

	if note.Failure != wantFailure {
		t.Fatalf("failure = %q, want %q", note.Failure, wantFailure)
	}

	if expect.key != "" && note.Key != expect.key {
		t.Fatalf("key = %q, want %q", note.Key, expect.key)
	}

	if expect.path != "" && note.Path != expect.path {
		t.Fatalf("path = %q, want %q", note.Path, expect.path)
	}

	if expect.hasVersion != note.HasVersion {
		t.Fatalf("HasVersion = %v, want %v", note.HasVersion, expect.hasVersion)
	}

	if expect.hasVersion && note.Version != expect.version {
		t.Fatalf("version = %d, want %d", note.Version, expect.version)
	}

	var buf bytes.Buffer

	diagnose.Write(&buf, note)

	text := buf.String()
	if expect.supported != "" && !strings.Contains(text, "supported versions: "+expect.supported) {
		t.Fatalf("diagnosis missing supported versions: %q", text)
	}

	if expect.omitSecret && strings.Contains(text, fixtureSecret) {
		t.Fatalf("diagnosis must not contain fixture secret")
	}

	if !strings.Contains(text, "aws-guard recover policy") {
		t.Fatalf("diagnosis missing recover command: %q", text)
	}

	if !expect.omitSecret {
		assertNoSecretInError(t, err)
	}
}

func configPath(t *testing.T, env []string) string {
	t.Helper()

	path, err := paths.ConfigFile(env)
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	return path
}
