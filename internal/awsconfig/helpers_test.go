package awsconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, home, configBody, credBody string) {
	t.Helper()

	awsDir := filepath.Join(home, ".aws")

	err := os.Mkdir(awsDir, 0o700)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if configBody != "" {
		writeFile(t, filepath.Join(awsDir, "config"), configBody)
	}

	if credBody != "" {
		writeFile(t, filepath.Join(awsDir, "credentials"), credBody)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()

	err := os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func assertNoSecretInError(t *testing.T, err error, secret string) {
	t.Helper()

	if err == nil {
		return
	}

	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaks secret: %v", err)
	}
}
