//go:build unix

package install_test

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDisableHeadlessRefuses(t *testing.T) { //nolint:paralleltest // headless Setsid subprocess
	prefix, _, disableScript := installPrefixForDisable(t)
	wrapperBefore := wrapperBytesAtAWS(t, prefix)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, disableScript, "disable")
	cmd.Dir = prefix
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err := cmd.Run()
	if exitStatus(err) != 3 {
		t.Fatalf("exit = %v, want 3", err)
	}

	if !strings.Contains(stderr.String(), "not confirmed") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	after := wrapperBytesAtAWS(t, prefix)
	if string(after) != string(wrapperBefore) {
		t.Fatal("aws must be unchanged")
	}
}

func TestDisableWrongPhraseRefuses(t *testing.T) { //nolint:paralleltest // PTY confirmation helper
	_, _, disableScript := installPrefixForDisable(t)
	wrapperBefore := wrapperBytesAtAWS(t, filepath.Dir(disableScript))

	combined, _, code := runDisableScript(t, disableScript, "disable", "wrong")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(combined, "not confirmed") {
		t.Fatalf("output = %q", combined)
	}

	after := wrapperBytesAtAWS(t, filepath.Dir(disableScript))
	if string(after) != string(wrapperBefore) {
		t.Fatal("aws must be unchanged")
	}
}
