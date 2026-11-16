package invocation_test

import (
	"aws-cli-guard/internal/invocation"
	"testing"
)

func TestParseCLIErrorFormatPreservesCredentialExport(t *testing.T) {
	t.Parallel()

	args := []string{
		awsBin,
		"--cli-error-format",
		"json",
		configureCmd,
		operationExportCreds,
	}

	got := invocation.Parse(args)
	if !got.OK {
		t.Fatalf("Parse() OK = false, reason %q", got.Reason)
	}

	if got.Command != configureCmd || got.Operation != operationExportCreds {
		t.Fatalf("command/operation = %q / %q", got.Command, got.Operation)
	}

	if !got.PrintsAccessKeys {
		t.Fatal("PrintsAccessKeys = false, want true")
	}
}

func TestParseCLIErrorFormatInvalidChoice(t *testing.T) {
	t.Parallel()

	got := invocation.Parse([]string{awsBin, "--cli-error-format", "bogus", "s3", "ls"})
	if got.OK {
		t.Fatal("Parse() OK = true, want false")
	}

	if got.Reason != reasonGlobalOptionInvalidValue {
		t.Fatalf("reason = %q, want %q", got.Reason, reasonGlobalOptionInvalidValue)
	}
}

func TestParseOutputOff(t *testing.T) {
	t.Parallel()

	got := invocation.Parse([]string{awsBin, "--output", "off", "s3", "ls"})
	if !got.OK {
		t.Fatalf("Parse() OK = false, reason %q", got.Reason)
	}

	if got.Command != "s3" || got.Operation != "ls" {
		t.Fatalf("command/operation = %q / %q", got.Command, got.Operation)
	}
}
