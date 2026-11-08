package sessions_test

import (
	"aws-cli-guard/internal/sessions"
	"strings"
	"testing"
)

func TestLoadOffWithoutTTY(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeSessions(t, env, `{
  "version": 1,
  "records": [
    {
      "session_id": "sid:1",
      "effect": "off",
      "token": "`+sentinelToken+`",
      "at": "2026-01-01T00:00:00Z"
    }
  ]
}
`)

	_, err := sessions.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadUnknownField(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeSessions(t, env, `{
  "version": 1,
  "records": [],
  "extra": true
}
`)

	_, err := sessions.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCorruptionOmitsToken(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeSessions(t, env, `{
  "version": 1,
  "records": [
    {
      "session_id": "sid:1",
      "effect": "off",
      "token": "not-a-token",
      "at": "2026-01-01T00:00:00Z",
      "tty": { "path": "/dev/tty", "dev": 1, "ino": 2, "rdev": 3 }
    }
  ]
}
`)

	_, err := sessions.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	note, ok := sessions.Corruption(err)
	if !ok {
		t.Fatal("Corruption = false")
	}

	out := note.Path + note.Failure + note.Key
	if strings.Contains(out, sentinelToken) {
		t.Fatal("diagnosis must not contain token material")
	}
}
