package awsconfig_test

import (
	"aws-cli-guard/internal/awsconfig"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigPathDefaults(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	env := []string{"HOME=" + home}

	got, err := awsconfig.ConfigPath(env)
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	want := filepath.Join(home, ".aws", "config")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCredentialsPathOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cred := filepath.Join(dir, "creds.ini")
	env := []string{
		"HOME=" + dir,
		"AWS_SHARED_CREDENTIALS_FILE=" + cred,
	}

	got, err := awsconfig.CredentialsPath(env)
	if err != nil {
		t.Fatalf("CredentialsPath: %v", err)
	}

	if got != cred {
		t.Fatalf("got %q, want %q", got, cred)
	}
}

func TestConfigPathEmptyOverrideUsesDefault(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	env := []string{
		"HOME=" + home,
		"AWS_CONFIG_FILE=",
	}

	got, err := awsconfig.ConfigPath(env)
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	want := filepath.Join(home, ".aws", "config")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPathExpandTildeAndVar(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	target := filepath.Join(home, "cfg", "config")
	env := []string{
		"HOME=" + home,
		"AWS_CONFIG_FILE=~/cfg/config",
	}

	got, err := awsconfig.ConfigPath(env)
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	if got != target {
		t.Fatalf("got %q, want %q", got, target)
	}

	env = []string{
		"HOME=" + home,
		"AWS_CONFIG_FILE=$HOME/other",
	}

	got, err = awsconfig.ConfigPath(env)
	if err != nil {
		t.Fatalf("ConfigPath var: %v", err)
	}

	want := filepath.Join(home, "other")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPathUnsetVarStaysLiteral(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	env := []string{
		"HOME=" + home,
		"AWS_CONFIG_FILE=$MISSING_DIR/file",
	}

	got, err := awsconfig.ConfigPath(env)
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	if !strings.Contains(got, "$MISSING_DIR") {
		t.Fatalf("expected literal unset var in %q", got)
	}
}

func TestPathHomeUnsetForDefault(t *testing.T) {
	t.Parallel()

	_, err := awsconfig.ConfigPath(nil)
	if !errors.Is(err, awsconfig.ErrHomeUnset) {
		t.Fatalf("errors.Is(err, ErrHomeUnset) = false, got %v", err)
	}
}

func TestPathTildeUserRejected(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	env := []string{
		"HOME=" + home,
		"AWS_CONFIG_FILE=~root/.aws/config",
	}

	_, err := awsconfig.ConfigPath(env)
	if !errors.Is(err, awsconfig.ErrTildeUser) {
		t.Fatalf("errors.Is(err, ErrTildeUser) = false, got %v", err)
	}
}
