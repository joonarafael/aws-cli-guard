package paths_test

import (
	"aws-cli-guard/internal/paths"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigFile(t *testing.T) {
	t.Parallel()

	for _, testCase := range configFileCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := copyEnv(testCase.env)
			got, err := paths.ConfigFile(env)

			if testCase.wantErr {
				assertConfigFileError(t, err)

				return
			}

			if err != nil {
				t.Fatalf("ConfigFile: %v", err)
			}

			if !filepath.IsAbs(got) {
				t.Fatalf("expected absolute path, got %q", got)
			}

			if testCase.want != "" && got != testCase.want {
				t.Fatalf("got %q, want %q", got, testCase.want)
			}

			if testCase.wantSuffix != "" &&
				!strings.HasSuffix(got, filepath.FromSlash(testCase.wantSuffix)) {
				t.Fatalf("unexpected path %q", got)
			}
		})
	}
}

func assertConfigFileError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("expected ErrHomeUnset, got %v", err)
	}
}

func TestSettingsFile(t *testing.T) {
	t.Parallel()

	env := copyEnv([]string{envXDGConfig, envHomeUser})

	got, err := paths.SettingsFile(env)
	if err != nil {
		t.Fatalf("SettingsFile: %v", err)
	}

	want := rootJoin("xdg", "config", "aws-cli-guard", "settings.toml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGrantsFile(t *testing.T) {
	t.Parallel()

	env := copyEnv([]string{envStateDir, envXDGState, envHomeUser})

	got, err := paths.GrantsFile(env)
	if err != nil {
		t.Fatalf("GrantsFile: %v", err)
	}

	want := rootJoin("state", "override", "grants.json")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCacheDir(t *testing.T) {
	t.Parallel()

	env := copyEnv([]string{envXDGCache, envHomeUser})

	got, err := paths.CacheDir(env)
	if err != nil {
		t.Fatalf("CacheDir: %v", err)
	}

	want := rootJoin("xdg", "cache", "aws-cli-guard")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPathsIgnoreProcessHome(t *testing.T) {
	t.Setenv("HOME", "/process/home/should/not/matter")

	env := []string{"XDG_CONFIG_HOME=/only/from/slice"}

	got, err := paths.ConfigFile(env)
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	want := rootJoin("only", "from", "slice", "aws-cli-guard", "config.toml")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func rootJoin(parts ...string) string {
	all := append([]string{string(filepath.Separator)}, parts...)

	return filepath.Join(all...)
}

func copyEnv(entries []string) []string {
	return append([]string(nil), entries...)
}
