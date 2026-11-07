//go:build unix

package install_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertInstallArtifacts(t *testing.T, prefix, wantRealAWS string, stdout *bytes.Buffer) {
	t.Helper()

	assertInstallFileCount(t, prefix)
	assertInstallModes(t, prefix)
	assertInstallManifest(t, prefix, wantRealAWS)
	assertInstallStdout(t, prefix, stdout)
}

func assertInstallFileCount(t *testing.T, prefix string) {
	t.Helper()

	entries, err := os.ReadDir(prefix)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 4 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}

		t.Fatalf("got %d install files %v, want 4", len(entries), names)
	}
}

func assertInstallModes(t *testing.T, prefix string) {
	t.Helper()

	manifestInfo, err := os.Stat(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	if manifestInfo.Mode().Perm() != 0o644 {
		t.Fatalf("manifest mode = %o, want 644", manifestInfo.Mode().Perm())
	}

	for _, name := range []string{archiveMemberAWS, archiveMemberGuard, archiveMemberDisable} {
		info, statErr := os.Stat(filepath.Join(prefix, name))
		if statErr != nil {
			t.Fatal(statErr)
		}

		if info.Mode().Perm() != 0o755 {
			t.Fatalf("%s mode = %o, want 755", name, info.Mode().Perm())
		}
	}

	for _, reserved := range []string{"aws.aside", "aws-cli-guard-disabled"} {
		_, statErr := os.Lstat(filepath.Join(prefix, reserved))
		if !os.IsNotExist(statErr) {
			t.Fatalf("reserved %s should be absent: %v", reserved, statErr)
		}
	}
}

func assertInstallManifest(t *testing.T, prefix, wantRealAWS string) {
	t.Helper()

	manifestBytes, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	wantLines := []string{
		"version=dev",
		"aws=" + filepath.Join(prefix, archiveMemberAWS),
		"aws_guard=" + filepath.Join(prefix, archiveMemberGuard),
		"real_aws=" + wantRealAWS,
		"disable_script=" + filepath.Join(prefix, archiveMemberDisable),
		"aside=" + filepath.Join(prefix, "aws.aside"),
		"marker=" + filepath.Join(prefix, "aws-cli-guard-disabled"),
	}

	manifest := string(manifestBytes)
	for _, line := range wantLines {
		if !strings.Contains(manifest, line) {
			t.Fatalf("manifest missing %q in:\n%s", line, manifest)
		}
	}
}

func assertInstallStdout(t *testing.T, prefix string, stdout *bytes.Buffer) {
	t.Helper()

	out := stdout.String()
	if !strings.Contains(out, "aws-cli-guard: prefix: "+prefix) {
		t.Fatalf("stdout = %q, want prefix detail", out)
	}

	if !strings.Contains(out, "export PATH='"+prefix+":$PATH'\n") {
		t.Fatalf("stdout = %q, want export line", out)
	}
}
