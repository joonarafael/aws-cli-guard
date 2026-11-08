package sessions_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const sentinelToken = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

func stateDirEnv(t *testing.T) []string {
	t.Helper()

	dir := t.TempDir()

	return []string{"AWS_CLI_GUARD_STATE_DIR=" + dir}
}

func writeSessions(t *testing.T, env []string, body string) {
	t.Helper()

	dir := envStateDir(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "sessions.json"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func envStateDir(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if len(entry) > len("AWS_CLI_GUARD_STATE_DIR=") &&
			entry[:len("AWS_CLI_GUARD_STATE_DIR=")] == "AWS_CLI_GUARD_STATE_DIR=" {
			return entry[len("AWS_CLI_GUARD_STATE_DIR="):]
		}
	}

	t.Fatal("missing state dir env")

	return ""
}

func sampleOffJSON(sessionID, ttyPath string, dev, ino, rdev int64) string {
	return `{
  "version": 1,
  "records": [
    {
      "session_id": "` + sessionID + `",
      "effect": "off",
      "token": "` + sentinelToken + `",
      "at": "2026-01-01T00:00:00Z",
      "tty": {
        "path": "` + ttyPath + `",
        "dev": ` + itoa64(dev) + `,
        "ino": ` + itoa64(ino) + `,
        "rdev": ` + itoa64(rdev) + `
      }
    }
  ]
}
`
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
