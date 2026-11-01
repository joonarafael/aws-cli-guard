package executor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	execSSOCommand = "sso"
	execLoginOp    = "login"
)

func assertSSOLoginRecordedArgs(t *testing.T, args []string) {
	t.Helper()

	if args[1] != execSSOCommand || args[2] != execLoginOp {
		t.Fatalf("unexpected command args: %v", args)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--sso-session") || !strings.Contains(joined, "dev") {
		t.Fatalf("recorded argv missing sso-session: %v", args)
	}

	if !strings.Contains(joined, "--profile") || !strings.Contains(joined, "default") {
		t.Fatalf("recorded argv missing profile: %v", args)
	}
}

func assertNoGrantsFile(t *testing.T, work string) {
	t.Helper()

	grantsPath := filepath.Join(work, "guard-state", "grants.json")

	_, statErr := os.Stat(grantsPath)
	if statErr == nil {
		t.Fatal("grants.json must not be created for login handoff")
	}

	if !os.IsNotExist(statErr) {
		t.Fatalf("Stat grants: %v", statErr)
	}
}
