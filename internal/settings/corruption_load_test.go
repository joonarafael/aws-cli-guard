package settings_test

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/settings"
	"testing"
)

func TestCorruptionFromLoadUnknownField(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writeSettings(t, env, `version = 1

[unknown]
value = true
`)

	_, err := settings.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	note, corrupt := settings.Corruption(err)
	if !corrupt {
		t.Fatal("Corruption = false, want true")
	}

	wantClass := diagnose.ClassSettings()
	if note.Failure != "unknown field" || note.Class.Name != wantClass.Name ||
		note.Class.Recover != wantClass.Recover {
		t.Fatalf("note = %+v", note)
	}
}
