//go:build unix

package install_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	archiveMemberAWS     = "aws"
	archiveMemberGuard   = "aws-guard"
	archiveMemberDisable = "aws-cli-guard-disable"
	linuxAmd64Archive    = "aws-cli-guard_1.2.3_linux_amd64.tar.gz"
)

func moduleRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func build(t *testing.T, dir, out, pkg string) string {
	t.Helper()

	cmd := exec.CommandContext(
		t.Context(),
		"go",
		"build",
		"-tags=awsguard_testpaths",
		"-o",
		out,
		pkg,
	)
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, output)
	}

	return out
}

func copyScript(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(dst, data, 0o755)
	if err != nil {
		t.Fatal(err)
	}
}

func stageBundle(t *testing.T, root string) string {
	t.Helper()

	bundle := filepath.Join(t.TempDir(), "bundle")

	err := os.MkdirAll(bundle, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	guard := build(t, root, filepath.Join(bundle, "aws-guard"), "./cmd/aws-guard")
	wrapperAWS := build(t, root, filepath.Join(bundle, "aws"), "./cmd/aws")
	disable := filepath.Join(bundle, "aws-cli-guard-disable")
	copyScript(t, filepath.Join(root, "scripts", "aws-cli-guard-disable"), disable)

	_ = wrapperAWS

	return guard
}

func buildFakeAWS(t *testing.T, root, dir string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	build(t, root, filepath.Join(dir, "aws"), "./internal/executor/testdata/fakeaws")
}
