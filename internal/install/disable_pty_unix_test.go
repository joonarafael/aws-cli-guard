//go:build unix

package install_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func runDisableScript(
	t *testing.T,
	scriptPath, action, phrase string,
) (string, string, int) {
	t.Helper()

	return runControlPlaneWithConfirm(t, scriptPath, []string{action}, phrase)
}

func runControlPlaneWithConfirm(
	t *testing.T,
	bin string,
	argv []string,
	phrase string,
) (string, string, int) {
	t.Helper()

	root := moduleRoot(t)
	runner := filepath.Join(root, "scripts", "install-pty-test.py")

	args := append([]string{runner, bin, phrase}, argv...)

	cmd := exec.CommandContext(t.Context(), "python3", append([]string{"-u"}, args...)...)
	cmd.Dir = filepath.Dir(bin)

	var combined bytes.Buffer

	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err := cmd.Run()

	return combined.String(), "", exitStatus(err)
}
