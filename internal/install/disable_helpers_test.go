//go:build unix

package install_test

import (
	"aws-cli-guard/internal/install"
	"bytes"
	"errors"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func installPrefixForDisable(t *testing.T) (string, string, string) {
	t.Helper()

	root := moduleRoot(t)
	work := t.TempDir()
	prefix := filepath.Join(work, "prefix")
	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)
	realAWS := filepath.Join(realDir, "aws")
	env := []string{"PATH=" + realDir}

	_, err := install.Run(testOpts(env, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	disableScript := filepath.Join(prefix, archiveMemberDisable)

	return prefix, realAWS, disableScript
}

func wrapperBytesAtAWS(t *testing.T, prefix string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(prefix, archiveMemberAWS))
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func runInstalledAWS(
	t *testing.T,
	prefix string,
	args ...string,
) (string, string, int) {
	t.Helper()

	cmd := execInstalledAWS(t, prefix, args...)

	var outBuf, errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	return outBuf.String(), errBuf.String(), exitStatus(err)
}

func execInstalledAWS(t *testing.T, prefix string, args ...string) *osexec.Cmd {
	t.Helper()

	bin := filepath.Join(prefix, archiveMemberAWS)

	return osexec.CommandContext(t.Context(), bin, args...)
}

func exitStatus(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := errors.AsType[*osexec.ExitError](err); ok {
		return exitErr.ExitCode()
	}

	return -1
}

func assertNoPathLeak(t *testing.T, realAWS, stdout, stderr string) {
	t.Helper()

	if strings.Contains(stdout, realAWS) || strings.Contains(stderr, realAWS) {
		t.Fatalf(
			"output must not contain real_aws path %q\nstdout=%q\nstderr=%q",
			realAWS,
			stdout,
			stderr,
		)
	}
}
