//go:build unix

//nolint:testpackage // exercises unexported minisign verification helpers.
package update

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"aead.dev/minisign"
)

func testMinisignBinary(t *testing.T) string {
	t.Helper()

	path, err := exec.LookPath("minisign")
	if err == nil {
		return path
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	root, err := moduleRoot(t)
	if err != nil {
		t.Fatal(err)
	}

	candidate := filepath.Join(root, ".cache", "minisign", "minisign-linux", arch, "minisign")

	st, statErr := os.Stat(candidate)
	if statErr == nil && !st.IsDir() {
		return candidate
	}

	t.Skip("minisign not on PATH and .cache/minisign binary missing")

	return ""
}

func moduleRoot(t *testing.T) (string, error) {
	t.Helper()

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err //nolint:wrapcheck // test helper keeps Getwd diagnostic.
	}

	for {
		_, statErr := os.Stat(filepath.Join(workingDir, "go.mod"))
		if statErr == nil {
			return workingDir, nil
		}

		parent := filepath.Dir(workingDir)
		if parent == workingDir {
			return "", os.ErrNotExist
		}

		workingDir = parent
	}
}

func generateMinisignKeypair(t *testing.T) (minisign.PublicKey, string) {
	t.Helper()

	work := t.TempDir()
	pubPath := filepath.Join(work, "test.pub")
	secPath := filepath.Join(work, "test.sec")

	cmd := exec.CommandContext(
		t.Context(),
		testMinisignBinary(t),
		"-G",
		"-s",
		secPath,
		"-p",
		pubPath,
		"-W",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("minisign -G: %v\n%s", err, out)
	}

	data, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatal(err)
	}

	var pub minisign.PublicKey

	err = pub.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}

	return pub, secPath
}

func signChecksumsCLI(t *testing.T, secPath string, checksums []byte) []byte {
	t.Helper()

	work := t.TempDir()
	checksumsPath := filepath.Join(work, "checksums.txt")

	err := os.WriteFile(checksumsPath, checksums, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(
		t.Context(),
		testMinisignBinary(t),
		"-Sm",
		checksumsPath,
		"-s",
		secPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("minisign -Sm: %v\n%s", err, out)
	}

	sig, err := os.ReadFile(checksumsPath + ".minisig")
	if err != nil {
		t.Fatal(err)
	}

	return sig
}

func TestEmbeddedPublicKeyLoads(t *testing.T) {
	t.Parallel()

	key, err := EmbeddedPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	if key.ID() == 0 {
		t.Fatal("expected non-zero key ID")
	}
}

func TestInstallPublicKeyMatchesEmbedded(t *testing.T) {
	t.Parallel()

	root, err := moduleRoot(t)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile(filepath.Join(root, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}

	pubFile, err := os.ReadFile(filepath.Join(root, "internal", "update", "minisign.pub"))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(pubFile)), "\n")
	want := strings.TrimSpace(lines[1])

	for line := range strings.SplitSeq(string(script), "\n") {
		if !strings.HasPrefix(line, "public_key='") {
			continue
		}

		got := strings.TrimPrefix(line, "public_key='")
		got = strings.TrimSuffix(got, "'")

		if got != want {
			t.Fatalf("install.sh public_key = %q, want %q", got, want)
		}

		return
	}

	t.Fatal("install.sh missing public_key line")
}

func TestVerifyChecksumsSignatureAcceptsMinisignCLI(t *testing.T) {
	t.Parallel()

	pub, secPath := generateMinisignKeypair(t)
	checksums := []byte("deadbeef  aws-cli-guard_1.0.0_linux_amd64.tar.gz\n")
	sig := signChecksumsCLI(t, secPath, checksums)

	err := verifyChecksumsSignature(pub, checksums, sig)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyChecksumsSignatureRejectsTamperedChecksums(t *testing.T) {
	t.Parallel()

	pub, secPath := generateMinisignKeypair(t)
	checksums := []byte("deadbeef  aws-cli-guard_1.0.0_linux_amd64.tar.gz\n")
	sig := signChecksumsCLI(t, secPath, checksums)

	tampered := bytes.ReplaceAll(checksums, []byte("dead"), []byte("beef"))

	err := verifyChecksumsSignature(pub, tampered, sig)
	if !errors.Is(err, errSignatureUntrusted) {
		t.Fatalf("err = %v, want %v", err, errSignatureUntrusted)
	}
}

func TestVerifyChecksumsSignatureRejectsWrongKey(t *testing.T) {
	t.Parallel()

	_, secPath := generateMinisignKeypair(t)
	otherPub, _ := generateMinisignKeypair(t)
	checksums := []byte("abcd  foo.tar.gz\n")
	sig := signChecksumsCLI(t, secPath, checksums)

	err := verifyChecksumsSignature(otherPub, checksums, sig)
	if !errors.Is(err, errSignatureInvalid) {
		t.Fatalf("err = %v, want %v", err, errSignatureInvalid)
	}
}

func TestVerifyChecksumsSignatureRejectsBadTrustedComment(t *testing.T) {
	t.Parallel()

	pub, secPath := generateMinisignKeypair(t)
	checksums := []byte("abcd  foo.tar.gz\n")
	sig := signChecksumsCLI(t, secPath, checksums)

	var parsed minisign.Signature

	err := parsed.UnmarshalText(sig)
	if err != nil {
		t.Fatal(err)
	}

	parsed.TrustedComment = "timestamp:1\tfile:other.txt\thashed"

	badSig, err := parsed.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	err = verifyChecksumsSignature(pub, checksums, badSig)
	if !errors.Is(err, errSignatureUntrusted) {
		t.Fatalf("err = %v, want %v", err, errSignatureUntrusted)
	}
}

func TestVerifyChecksumsSignatureRejectsOversizedSignature(t *testing.T) {
	t.Parallel()

	pub, _ := generateMinisignKeypair(t)
	checksums := []byte("abcd\n")
	huge := make([]byte, maxSignatureSize+1)

	err := verifyChecksumsSignature(pub, checksums, huge)
	if !errors.Is(err, errSignatureTooLarge) {
		t.Fatalf("err = %v, want %v", err, errSignatureTooLarge)
	}
}
