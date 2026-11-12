//go:build unix

package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDisableEnableRoundTrip(t *testing.T) { //nolint:paralleltest // PTY confirmation helper
	prefix, realAWS, disableScript := installPrefixForDisable(t)
	wrapperBefore := wrapperBytesAtAWS(t, prefix)

	assertDisableSucceeds(t, disableScript, prefix, wrapperBefore)
	assertStandInRunsFakeAWS(t, prefix, realAWS)
	assertEnableRestoresWrapper(t, disableScript, prefix, wrapperBefore)
}

func assertDisableSucceeds(t *testing.T, disableScript, prefix string, wrapperBefore []byte) {
	t.Helper()

	out, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 0 {
		t.Fatalf("disable exit = %d, want 0; output=%q", code, out)
	}

	if !strings.Contains(out, "this installation is disabled and policy is not applied") {
		t.Fatalf("disable stdout = %q", out)
	}

	asideData, err := os.ReadFile(filepath.Join(prefix, "aws.aside"))
	if err != nil {
		t.Fatal(err)
	}

	if string(asideData) != string(wrapperBefore) {
		t.Fatal("wrapper bytes should live at aws.aside")
	}

	_, err = os.Stat(filepath.Join(prefix, "aws-cli-guard-disabled"))
	if err != nil {
		t.Fatal(err)
	}
}

func assertStandInRunsFakeAWS(t *testing.T, prefix, realAWS string) {
	t.Helper()

	stdout, stderr, awsCode := runInstalledAWS(t, prefix, "--version")
	if awsCode != 0 {
		t.Fatalf("aws exit = %d", awsCode)
	}

	if !strings.Contains(stderr, warnDisabled) {
		t.Fatalf("stderr = %q, want warning", stderr)
	}

	if !strings.Contains(stdout, "aws-cli/2.9.0") {
		t.Fatalf("stdout = %q", stdout)
	}

	assertNoPathLeak(t, realAWS, stdout, stderr)
}

func assertEnableRestoresWrapper(t *testing.T, disableScript, prefix string, wrapperBefore []byte) {
	t.Helper()

	out, _, code := runDisableScript(t, disableScript, "enable", "enable")
	if code != 0 {
		t.Fatalf("enable exit = %d", code)
	}

	if !strings.Contains(out, "this installation is enabled and policy applies") {
		t.Fatalf("enable stdout = %q", out)
	}

	restored := wrapperBytesAtAWS(t, prefix)
	if string(restored) != string(wrapperBefore) {
		t.Fatal("wrapper should be restored at aws")
	}

	_, err := os.Lstat(filepath.Join(prefix, "aws.aside"))
	if !os.IsNotExist(err) {
		t.Fatal("aside should be gone")
	}

	_, err = os.Lstat(filepath.Join(prefix, "aws-cli-guard-disabled"))
	if !os.IsNotExist(err) {
		t.Fatal("marker should be gone")
	}
}
