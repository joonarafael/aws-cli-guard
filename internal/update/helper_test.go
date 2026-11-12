//go:build unix

package update_test

import (
	"aws-cli-guard/internal/update"
	"testing"
)

func TestReleaseStamp(t *testing.T) {
	t.Parallel()

	if !update.ReleaseStamp("1.2.3") {
		t.Fatal("expected release stamp")
	}

	if update.ReleaseStamp("dev") {
		t.Fatal("dev must not check")
	}
}

func TestSupportedPlatform(t *testing.T) {
	t.Parallel()

	if !update.SupportedPlatform("linux", "amd64") {
		t.Fatal("linux amd64 should be supported")
	}

	if update.SupportedPlatform("windows", "amd64") {
		t.Fatal("windows should be unsupported")
	}
}
