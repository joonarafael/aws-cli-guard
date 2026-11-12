package verbose_test

import (
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/verbose"
	"errors"
	"testing"
)

func TestEnvLevelUnset(t *testing.T) {
	t.Parallel()

	got, err := verbose.EnvLevel(nil)
	if err != nil || got != settings.LevelTight {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestEnvLevelEmpty(t *testing.T) {
	t.Parallel()

	got, err := verbose.EnvLevel([]string{"AWS_CLI_GUARD_LOG="})
	if err != nil || got != settings.LevelTight {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestEnvLevelVerbose(t *testing.T) {
	t.Parallel()

	got, err := verbose.EnvLevel([]string{"AWS_CLI_GUARD_LOG=verbose"})
	if err != nil || got != settings.LevelVerbose {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestEnvLevelInvalidNoEcho(t *testing.T) {
	t.Parallel()

	secret := "super-secret-level"

	_, err := verbose.EnvLevel([]string{"AWS_CLI_GUARD_LOG=" + secret})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, verbose.ErrInvalidEnv) {
		t.Fatalf("got %v", err)
	}
}

func TestEffectiveEnvBeatsFile(t *testing.T) {
	t.Parallel()

	got, err := verbose.Effective(
		[]string{"AWS_CLI_GUARD_LOG=tight"},
		settings.LevelVerbose,
	)
	if err != nil || got != settings.LevelTight {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestEffectiveUsesFileWhenEnvUnset(t *testing.T) {
	t.Parallel()

	got, err := verbose.Effective(nil, settings.LevelVerbose)
	if err != nil || got != settings.LevelVerbose {
		t.Fatalf("got %q, %v", got, err)
	}
}
