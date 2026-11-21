//go:build unix

package install_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func minisignBinary(t *testing.T) string {
	t.Helper()

	path, err := exec.LookPath("minisign")
	if err == nil {
		return path
	}

	arch := runtime.GOARCH

	const archAMD64 = "amd64"

	if arch == archAMD64 {
		arch = "x86_64"
	}

	root := moduleRoot(t)
	candidate := filepath.Join(
		root,
		".cache",
		"minisign",
		"minisign-linux",
		arch,
		"minisign",
	)

	st, statErr := os.Stat(candidate)
	if statErr == nil && !st.IsDir() {
		return candidate
	}

	t.Skip("minisign not on PATH and .cache/minisign binary missing")

	return ""
}

func generateMinisignKeypair(t *testing.T) (string, string) {
	t.Helper()

	work := t.TempDir()
	pubPath := filepath.Join(work, "test.pub")
	secPath := filepath.Join(work, "test.sec")

	cmd := exec.CommandContext(
		t.Context(),
		minisignBinary(t),
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

	pubB64 := minisignPublicKeyBase64(t, pubPath)

	return secPath, pubB64
}

func minisignPublicKeyBase64(t *testing.T, pubPath string) string {
	t.Helper()

	data, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatal("minisign public key file too short")
	}

	return strings.TrimSpace(lines[1])
}

func signChecksums(t *testing.T, secPath string, checksums []byte) []byte {
	t.Helper()

	work := t.TempDir()
	checksumsPath := filepath.Join(work, "checksums.txt")

	err := os.WriteFile(checksumsPath, checksums, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(
		t.Context(),
		minisignBinary(t),
		"-Sm",
		checksumsPath,
		"-s",
		secPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("minisign -Sm: %v\n%s", err, out)
	}

	sigPath := checksumsPath + ".minisig"

	sig, err := os.ReadFile(sigPath)
	if err != nil {
		t.Fatal(err)
	}

	return sig
}

var installPublicKeyRE = regexp.MustCompile(`(?m)^public_key='[^']*'`)

func substituteInstallPublicKey(script []byte, pubB64 string) []byte {
	replacement := fmt.Sprintf("public_key='%s'", pubB64)

	return installPublicKeyRE.ReplaceAll(script, []byte(replacement))
}

func installPublicKeyLine(script []byte) string {
	match := installPublicKeyRE.FindSubmatch(script)
	if match == nil {
		return ""
	}

	line := string(match[0])

	const prefix = "public_key='"

	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, "'") {
		return ""
	}

	return line[len(prefix) : len(line)-1]
}

func writeInstallScriptCopy(t *testing.T, root, pubB64 string) string {
	t.Helper()

	scriptPath := filepath.Join(root, "install.sh")

	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal(err)
	}

	copyPath := filepath.Join(t.TempDir(), "install.sh")

	err = os.WriteFile(copyPath, substituteInstallPublicKey(data, pubB64), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	return copyPath
}

func pathPrepend(dir string) string {
	sep := string(os.PathListSeparator)

	return dir + sep + os.Getenv("PATH")
}

func installScriptEnvWithMinisign(realBinDir, minisignDir string) []string {
	sep := string(os.PathListSeparator)
	path := realBinDir + sep + minisignDir + sep + "/usr/bin" + sep + "/bin"
	env := make([]string, 0, len(os.Environ())+1)

	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "PATH=") {
			continue
		}

		env = append(env, entry)
	}

	env = append(env, "PATH="+path)

	return env
}

func assertInstallKeyMatchesRepo(t *testing.T, root string) {
	t.Helper()

	script, err := os.ReadFile(filepath.Join(root, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}

	pubFile, err := os.ReadFile(filepath.Join(root, "internal", "update", "minisign.pub"))
	if err != nil {
		t.Fatal(err)
	}

	want := minisignPublicKeyBase64FromBytes(t, pubFile)
	got := installPublicKeyLine(script)

	if got != want {
		t.Fatalf("install.sh public_key = %q, want %q", got, want)
	}
}

func minisignPublicKeyBase64FromBytes(t *testing.T, data []byte) string {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatal("minisign public key file too short")
	}

	return strings.TrimSpace(lines[1])
}

func minisignDir(t *testing.T) string {
	t.Helper()

	bin := minisignBinary(t)

	return filepath.Dir(bin)
}
