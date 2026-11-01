package executor_test

import (
	"aws-cli-guard/internal/executor"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runWrapperVersionCase(
	t *testing.T,
	work string,
	wrapper string,
	fake string,
	extraEnv []string,
	args ...string,
) (string, error) {
	t.Helper()

	record := filepath.Join(work, "record.json")
	versionCount := filepath.Join(work, "version-count")

	cmd := exec.CommandContext(t.Context(), wrapper, args...)

	var stderrBuf strings.Builder

	cmd.Stderr = &stderrBuf
	cmd.Env = envHandoff(append(
		extraEnv,
		executor.RealAWSEnv+"="+fake,
		"FAKE_AWS_RECORD="+record,
		"FAKE_AWS_VERSION_COUNT="+versionCount,
	)...)

	err := cmd.Run()

	return stderrBuf.String(), err
}

func readVersionProbeCount(t *testing.T, work string) int {
	t.Helper()

	path := filepath.Join(work, "version-count")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}

		t.Fatal(err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	return count
}

func versionHandoffEnv(t *testing.T, work string, versionLine string) []string {
	t.Helper()

	base := policyConfigEnvEntries(t, work)
	if versionLine == "" {
		return base
	}

	return append(base, "FAKE_AWS_VERSION="+versionLine)
}

func assertCancellationWithOptionalWarning(t *testing.T, stderr string) {
	t.Helper()

	if !strings.Contains(stderr, "aws-cli-guard: error: command was not executed") {
		t.Fatalf("missing cancellation line in stderr: %q", stderr)
	}

	if !strings.Contains(stderr, "aws-cli-guard: error: AWS CLI Guard stopped it:") {
		t.Fatalf("missing stopped-it line in stderr: %q", stderr)
	}

	if strings.Contains(stderr, cancellationSecretArg) {
		t.Fatalf("stderr must not contain argv token")
	}
}

func writeCorruptCLIVersionCache(t *testing.T, cacheDir string) {
	t.Helper()

	err := os.MkdirAll(cacheDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(cacheDir, "cli-version.json"), []byte("{bad"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
