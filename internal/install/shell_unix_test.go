//go:build unix

package install_test

import (
	"aws-cli-guard/internal/install"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyShellCreatesBashrc(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	var stdout bytes.Buffer

	result, err := install.ApplyShell(
		shellOpts(env, home, prefix, install.ShellBash, false),
		&stdout,
	)
	if err != nil {
		t.Fatalf("ApplyShell: %v", err)
	}

	if !result.Changed {
		t.Fatal("want changed")
	}

	data, err := os.ReadFile(filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), prefix) {
		t.Fatalf("bashrc = %q", data)
	}

	if !strings.Contains(stdout.String(), "updated shell startup") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestApplyShellPreservesMode(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	rcPath := filepath.Join(home, ".bashrc")

	err := os.WriteFile(rcPath, []byte("# custom\n"), 0o640)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.ApplyShell(
		shellOpts(env, home, prefix, install.ShellBash, false),
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("ApplyShell: %v", err)
	}

	info, err := os.Stat(rcPath)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o, want 640", info.Mode().Perm())
	}
}

func TestApplyShellNotInstalledPrefix(t *testing.T) {
	t.Parallel()

	_, home, env := shellTestHome(t)

	_, err := install.ApplyShell(
		shellOpts(env, home, filepath.Join(home, "missing-install"), install.ShellBash, false),
		&bytes.Buffer{},
	)
	if !errors.Is(err, install.ErrNotInstalled) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyShellDisableLayoutRefuses(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	err := os.WriteFile(filepath.Join(prefix, "aws.aside"), []byte("x"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.ApplyShell(
		shellOpts(env, home, prefix, install.ShellBash, false),
		&bytes.Buffer{},
	)
	if !errors.Is(err, install.ErrDisableLayout) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyShellSymlinkRefuses(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	target := filepath.Join(home, "real-rc")

	err := os.WriteFile(target, []byte("# real\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(target, filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.ApplyShell(
		shellOpts(env, home, prefix, install.ShellBash, false),
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApplyShellRemove(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	var stdout bytes.Buffer

	_, err := install.ApplyShell(shellOpts(env, home, prefix, install.ShellZsh, false), &stdout)
	if err != nil {
		t.Fatalf("ApplyShell: %v", err)
	}

	_, err = install.ApplyShell(shellOpts(env, home, prefix, install.ShellZsh, true), &stdout)
	if err != nil {
		t.Fatalf("ApplyShell remove: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "aws-cli-guard") {
		t.Fatalf("zshrc = %q", data)
	}
}

func TestApplyShellTooLarge(t *testing.T) {
	t.Parallel()

	prefix, home, env := shellTestHome(t)

	large := make([]byte, (1<<20)+1)
	for i := range large {
		large[i] = 'a'
	}

	err := os.WriteFile(filepath.Join(home, ".bashrc"), large, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.ApplyShell(
		shellOpts(env, home, prefix, install.ShellBash, false),
		&bytes.Buffer{},
	)
	if !errors.Is(err, install.ErrShellStartupTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

func TestShellReasonMapsMarkers(t *testing.T) {
	t.Parallel()

	want := "the shell startup file has conflicting marker comments"

	got := install.Reason(install.ErrShellMarkersConflict)
	if got != want {
		t.Fatalf("Reason = %q, want %q", got, want)
	}
}

func shellTestHome(t *testing.T) (string, string, []string) {
	t.Helper()

	root := moduleRoot(t)
	work := t.TempDir()
	home := filepath.Join(work, "home")
	prefix := filepath.Join(work, "install")

	err := os.MkdirAll(home, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	guard := stageBundle(t, root)
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	env := []string{"PATH=" + realDir, "HOME=" + home}

	_, err = install.Run(testOpts(env, prefix, guard), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	return prefix, home, env
}

func shellOpts(
	env []string,
	home, prefix string,
	shell install.ShellName,
	remove bool,
) install.ShellOptions {
	return install.ShellOptions{
		Env:    env,
		Home:   home,
		Prefix: prefix,
		Shell:  shell,
		Remove: remove,
	}
}
