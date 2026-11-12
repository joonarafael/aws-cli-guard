package cli_test

import (
	"aws-cli-guard/internal/cli"
	"aws-cli-guard/internal/output"
	"bytes"
	"strings"
	"testing"
)

func TestRunDisableExtraArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, "disable", cliExtraArg}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	if !strings.Contains(stderr.String(), "usage: aws-guard disable") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunEnableExtraArgs(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	status := cli.Run([]string{progName, "enable", cliExtraArg}, nil, &bytes.Buffer{}, &stderr)
	if status != output.CancelStatus {
		t.Fatalf("status = %d, want %d", status, output.CancelStatus)
	}

	if !strings.Contains(stderr.String(), "usage: aws-guard enable") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
