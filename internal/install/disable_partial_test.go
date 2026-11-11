//go:build unix

package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//nolint:paralleltest // PTY confirmation helper
func TestDisablePartialAsideAWSNoMarker(t *testing.T) {
	prefix, _, disableScript := installPrefixForDisable(t)

	err := os.Rename(filepath.Join(prefix, archiveMemberAWS), filepath.Join(prefix, "aws.aside"))
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(
		filepath.Join(prefix, archiveMemberAWS),
		[]byte("#!/bin/sh\necho stand-in\n"),
		0o755,
	)
	if err != nil {
		t.Fatal(err)
	}

	out, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	if !strings.Contains(out, "disabled and policy is not applied") {
		t.Fatalf("stdout = %q", out)
	}

	_, err = os.Stat(filepath.Join(prefix, "aws-cli-guard-disabled"))
	if err != nil {
		t.Fatal(err)
	}
}

//nolint:paralleltest // PTY confirmation helper
func TestDisablePartialAsideNoAWSNoIncoming(t *testing.T) {
	prefix, _, disableScript := installPrefixForDisable(t)

	err := os.Rename(filepath.Join(prefix, archiveMemberAWS), filepath.Join(prefix, "aws.aside"))
	if err != nil {
		t.Fatal(err)
	}

	combined, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(combined, "run enable to restore") {
		t.Fatalf("output = %q", combined)
	}
}

//nolint:paralleltest // PTY confirmation helper
func TestDisableMarkerWithoutAsideRefuses(t *testing.T) {
	prefix, _, disableScript := installPrefixForDisable(t)

	err := os.WriteFile(filepath.Join(prefix, "aws-cli-guard-disabled"), nil, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	combined, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(combined, "layout cannot be changed") {
		t.Fatalf("output = %q", combined)
	}
}

//nolint:paralleltest // PTY confirmation helper
func TestDisableSymlinkAWSRefuses(t *testing.T) {
	prefix, _, disableScript := installPrefixForDisable(t)
	awsPath := filepath.Join(prefix, archiveMemberAWS)

	err := os.Rename(awsPath, awsPath+".bak")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(awsPath+".bak", awsPath)
	if err != nil {
		t.Fatal(err)
	}

	combined, _, code := runDisableScript(t, disableScript, "disable", "disable")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(combined, "layout cannot be changed") {
		t.Fatalf("output = %q", combined)
	}
}

//nolint:paralleltest // PTY confirmation helper
func TestStandInMissingRealAWS(t *testing.T) {
	prefix, realAWS, disableScript := installPrefixForDisable(t)

	runDisableScript(t, disableScript, "disable", "disable")

	err := os.Remove(realAWS)
	if err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runInstalledAWS(t, prefix, "--version")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(stderr, "recorded AWS CLI is missing") {
		t.Fatalf("stderr = %q", stderr)
	}
}

//nolint:paralleltest // PTY confirmation helper
func TestStandInRejectsRealAWSAtInstallAWS(t *testing.T) {
	prefix, _, disableScript := installPrefixForDisable(t)
	manifestPath := filepath.Join(prefix, "aws-cli-guard-manifest")

	runDisableScript(t, disableScript, "disable", "disable")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "real_aws=") {
			lines[i] = "real_aws=" + filepath.Join(prefix, archiveMemberAWS)

			break
		}
	}

	err = os.WriteFile(manifestPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runInstalledAWS(t, prefix, "--version")
	if code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}

	if !strings.Contains(stderr, "recorded AWS CLI is missing") {
		t.Fatalf("stderr = %q", stderr)
	}
}
