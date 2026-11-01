package executor_test

import (
	"aws-cli-guard/internal/executor"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWrapperSSOLoginWithoutPolicyExecs(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		execSSOCommand,
		execLoginOp,
		"--sso-session",
		"dev",
		"--profile",
		"default",
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	home := filepath.Join(work, "home")

	err := os.MkdirAll(home, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	cmd.Env = envHandoff(append(
		guardDirEnvEntries(t, work),
		"HOME="+home,
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
	)...)

	runErr := cmd.Run()
	if runErr != nil {
		t.Fatalf("run wrapper: %v", runErr)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty wrapper stderr on login handoff, got %q", stderr.String())
	}

	assertSSOLoginRecordedArgs(t, readRecordedArgs(t, record))
	assertNoGrantsFile(t, work)
}
