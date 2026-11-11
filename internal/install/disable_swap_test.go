//go:build unix

package install_test

import (
	"aws-cli-guard/internal/install"
	"bytes"
	"errors"
	"strings"
	"testing"
)

const warnDisabled = "aws-cli-guard: warning: this installation is disabled" +
	" and policy is not applied"

func TestDisableSecondCallNoOp(t *testing.T) { //nolint:paralleltest // PTY confirmation helper
	_, _, disableScript := installPrefixForDisable(t)

	_, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 0 {
		t.Fatalf("first disable exit = %d", code)
	}

	out, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 0 {
		t.Fatalf("second disable exit = %d", code)
	}

	if !strings.Contains(out, "already disabled") {
		t.Fatalf("stdout = %q", out)
	}
}

func TestEnableSecondCallNoOp(t *testing.T) { //nolint:paralleltest // PTY confirmation helper
	_, _, disableScript := installPrefixForDisable(t)

	runDisableScript(t, disableScript, "disable", "disable")

	_, _, code := runDisableScript(t, disableScript, "enable", "enable")
	if code != 0 {
		t.Fatalf("enable exit = %d", code)
	}

	out, _, code := runDisableScript(t, disableScript, "enable", "enable")
	if code != 0 {
		t.Fatalf("second enable exit = %d", code)
	}

	if !strings.Contains(out, "already enabled") {
		t.Fatalf("stdout = %q", out)
	}
}

func TestUpgradeRefusesAfterDisable(t *testing.T) { //nolint:paralleltest // PTY confirmation helper
	root := moduleRoot(t)
	prefix, _, disableScript := installPrefixForDisable(t)

	runDisableScript(t, disableScript, "disable", "disable")

	_, err := install.Upgrade(testOpts(nil, prefix, stageBundle(t, root)), &bytes.Buffer{})
	if !errors.Is(err, install.ErrDisableLayout) {
		t.Fatalf("Upgrade err = %v, want ErrDisableLayout", err)
	}
}
