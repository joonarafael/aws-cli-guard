package executor_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWrapperCorruptSettingsUnchanged(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	settingsPath := writeCorruptSettingsFixture(t, work)

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	settingsBody, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	cmd := exec.CommandContext(t.Context(), wrapper, "sts", cancellationSecretArg)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err = cmd.Run()
	assertExitCode(t, err)

	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}

	if string(after) != string(settingsBody) {
		t.Fatal("corrupt settings bytes changed")
	}

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run when settings are corrupt")
	}
}

func writeCorruptSettingsFixture(t *testing.T, work string) string {
	t.Helper()

	configDir := filepath.Join(work, "guard-config")
	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	writeEnforcePolicy(t, work, `version = 1

[defaults]
action = "prompt"

[credential_export]
action = "prompt"
`)

	settingsPath := filepath.Join(configDir, "settings.toml")

	err = os.WriteFile(settingsPath, []byte("not valid toml\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return settingsPath
}
