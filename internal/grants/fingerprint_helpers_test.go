package grants_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	fixtureCredProcFP = "UNIQUE-CRED-PROC-COMMAND-GR2-TEST"
	fixtureRoleARN    = "arn:aws:iam::123456789012:role/Admin"
)

func homeEnv(t *testing.T) []string {
	t.Helper()

	return []string{"HOME=" + t.TempDir()}
}

func writeAWSConfig(t *testing.T, env []string, configBody, credBody string) {
	t.Helper()

	home := homeFromEnv(t, env)
	awsDir := filepath.Join(home, ".aws")

	err := os.Mkdir(awsDir, 0o700)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if configBody != "" {
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configBody), 0o600)
		if err != nil {
			t.Fatalf("write config: %v", err)
		}
	}

	if credBody != "" {
		err = os.WriteFile(filepath.Join(awsDir, "credentials"), []byte(credBody), 0o600)
		if err != nil {
			t.Fatalf("write credentials: %v", err)
		}
	}
}

func homeFromEnv(t *testing.T, env []string) string {
	t.Helper()

	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, "HOME="); ok {
			return after
		}
	}

	t.Fatal("HOME not in env")

	return ""
}
