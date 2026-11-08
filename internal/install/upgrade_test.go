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

func TestUpgradeRefusesEmptyPrefix(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	prefix := filepath.Join(t.TempDir(), "empty")

	_, err := install.Upgrade(testOpts(nil, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if !errors.Is(err, install.ErrNotInstalled) {
		t.Fatalf("Upgrade err = %v, want ErrNotInstalled", err)
	}
}

func TestUpgradeRefusesAsidePresent(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	env := []string{"PATH=" + realDir}

	_, err := install.Run(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	err = os.WriteFile(filepath.Join(prefix, "aws.aside"), []byte("x"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.Upgrade(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if !errors.Is(err, install.ErrDisableLayout) {
		t.Fatalf("Upgrade err = %v, want ErrDisableLayout", err)
	}
}

func TestUpgradeRefusesSymlinkAWS(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	env := []string{"PATH=" + realDir}

	guard := stageBundle(t, root)

	_, err := install.Run(testOpts(env, prefix, guard), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	wrapperAWS := filepath.Join(prefix, "aws")

	data, err := os.ReadFile(wrapperAWS)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(wrapperAWS)
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(work, "target-aws")

	err = os.WriteFile(target, data, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(target, wrapperAWS)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.Upgrade(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if !errors.Is(err, install.ErrNotOurInstallation) {
		t.Fatalf("Upgrade err = %v, want ErrNotOurInstallation", err)
	}
}

func TestUpgradeSameBinaryPathSkipsSelfTruncate(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	env := []string{"PATH=" + realDir}

	guard := stageBundle(t, root)

	_, err := install.Run(testOpts(env, prefix, guard), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	installedGuard := filepath.Join(prefix, "aws-guard")

	before, err := os.ReadFile(installedGuard)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.Upgrade(testOpts(env, prefix, installedGuard), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Upgrade from installed guard: %v", err)
	}

	after, err := os.ReadFile(installedGuard)
	if err != nil {
		t.Fatal(err)
	}

	if len(after) == 0 || len(after) < len(before)/2 {
		t.Fatalf("guard truncated: before %d after %d", len(before), len(after))
	}
}

func TestUpgradeInvalidManifestOutsidePrefix(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	env := []string{"PATH=" + realDir}

	_, err := install.Run(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	manifestPath := filepath.Join(prefix, "aws-cli-guard-manifest")

	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	bad := strings.Replace(
		string(body),
		"aws="+filepath.Join(prefix, "aws"),
		"aws=/tmp/not-in-prefix/aws",
		1,
	)

	err = os.WriteFile(manifestPath, []byte(bad), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = install.Upgrade(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if !errors.Is(err, install.ErrManifestInvalid) {
		t.Fatalf("Upgrade err = %v, want ErrManifestInvalid", err)
	}
}
