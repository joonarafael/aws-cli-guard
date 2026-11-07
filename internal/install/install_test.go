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

func TestInstallSuccess(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")

	guard := stageBundle(t, root)
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	env := []string{"PATH=" + realDir}

	var stdout bytes.Buffer

	err := install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: guard,
	}, &stdout)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	assertInstallArtifacts(t, prefix, filepath.Join(realDir, "aws"), &stdout)
}

func TestInstallSkipsEarlierWrapperOnPATH(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")

	guard := stageBundle(t, root)
	wrapperAWS := filepath.Join(filepath.Dir(guard), archiveMemberAWS)
	extraWrapperDir := filepath.Join(work, "extra-wrapper")
	realDir := filepath.Join(work, "real-bin")

	err := os.MkdirAll(extraWrapperDir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(wrapperAWS)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(extraWrapperDir, "aws"), data, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	buildFakeAWS(t, root, realDir)

	pathEnv := extraWrapperDir + string(os.PathListSeparator) + realDir
	env := []string{"PATH=" + pathEnv}

	err = install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	manifest, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	wantReal := filepath.Join(realDir, "aws")
	if !strings.Contains(string(manifest), "real_aws="+wantReal) {
		t.Fatalf("manifest = %q, want real_aws=%q", manifest, wantReal)
	}
}

func TestInstallIgnoresRealAWSEnv(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")

	guard := stageBundle(t, root)
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	decoyDir := filepath.Join(work, "decoy")
	buildFakeAWS(t, root, decoyDir)

	env := []string{
		"PATH=" + realDir,
		"AWS_CLI_GUARD_REAL_AWS=" + filepath.Join(decoyDir, "aws"),
	}

	err := install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	manifest, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	want := "real_aws=" + filepath.Join(realDir, "aws")
	if !strings.Contains(string(manifest), want) {
		t.Fatalf("manifest = %q, want %q", manifest, want)
	}
}

func TestInstallRefusesOccupiedAWS(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")

	err := os.MkdirAll(prefix, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(prefix, "aws"), []byte("taken"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	guard := stageBundle(t, root)
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	err = install.Run(install.Options{
		Env:        []string{"PATH=" + realDir},
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if !errors.Is(err, install.ErrPathOccupied) {
		t.Fatalf("Run err = %v, want ErrPathOccupied", err)
	}

	_, err = os.Stat(filepath.Join(prefix, "aws-guard"))
	if !os.IsNotExist(err) {
		t.Fatalf("aws-guard should not be created: %v", err)
	}
}

func TestInstallRefusesMissingSibling(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")
	bundle := filepath.Join(work, "bundle")

	err := os.MkdirAll(bundle, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	guard := build(t, root, filepath.Join(bundle, "aws-guard"), "./cmd/aws-guard")
	build(t, root, filepath.Join(bundle, "aws"), "./cmd/aws")

	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	err = install.Run(install.Options{
		Env:        []string{"PATH=" + realDir},
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if !errors.Is(err, install.ErrSiblingMissing) {
		t.Fatalf("Run err = %v, want ErrSiblingMissing", err)
	}
}

func TestInstallRefusesInvalidPrefix(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	guard := stageBundle(t, root)
	realDir := filepath.Join(t.TempDir(), "real-bin")
	buildFakeAWS(t, root, realDir)

	err := install.Run(install.Options{
		Env:        []string{"PATH=" + realDir},
		Prefix:     "/tmp/bad'prefix",
		Executable: guard,
	}, &bytes.Buffer{})
	if !errors.Is(err, install.ErrPrefixInvalid) {
		t.Fatalf("Run err = %v, want ErrPrefixInvalid", err)
	}
}

func TestInstallSecondRunRefuses(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "install")

	guard := stageBundle(t, root)
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	env := []string{"PATH=" + realDir}

	err := install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}

	firstManifest, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	err = install.Run(install.Options{
		Env:        env,
		Prefix:     prefix,
		Executable: guard,
	}, &bytes.Buffer{})
	if !errors.Is(err, install.ErrPathOccupied) {
		t.Fatalf("second Run err = %v, want ErrPathOccupied", err)
	}

	secondManifest, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	if string(firstManifest) != string(secondManifest) {
		t.Fatal("manifest changed on refused second install")
	}
}
