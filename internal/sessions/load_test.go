package sessions_test

import (
	"aws-cli-guard/internal/sessions"
	"testing"
)

func TestLoadMissingFileEmpty(t *testing.T) {
	t.Parallel()

	got, err := sessions.Load(stateDirEnv(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Records) != 0 {
		t.Fatalf("records = %d", len(got.Records))
	}
}

func TestLoadOffRecord(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeSessions(t, env, sampleOffJSON("sid:42", "/dev/pts/7", 5, 9, 8))

	got, err := sessions.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Records) != 1 || got.Records[0].Effect != sessions.EffectOff {
		t.Fatalf("records: %+v", got.Records)
	}
}
