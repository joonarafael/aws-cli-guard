//go:build unix

package install_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func runReleaseAssets(t *testing.T, root, version, outdir string) error {
	t.Helper()

	script := filepath.Join(root, "scripts", "release-assets.sh")
	cmd := exec.CommandContext(t.Context(), "sh", script, version, outdir)
	cmd.Dir = root

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, output)
	}

	return nil
}

func assertTarMembers(t *testing.T, archivePath string) {
	t.Helper()

	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		closeErr := gzipReader.Close()
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}()

	tarReader := tar.NewReader(gzipReader)

	var names []string

	for {
		hdr, readErr := tarReader.Next()
		if errors.Is(readErr, io.EOF) {
			break
		}

		if readErr != nil {
			t.Fatal(readErr)
		}

		names = append(names, hdr.Name)
	}

	want := []string{archiveMemberAWS, archiveMemberGuard, archiveMemberDisable}
	if len(names) != len(want) {
		t.Fatalf("tar members = %v, want %v", names, want)
	}

	for i, name := range want {
		if names[i] != name {
			t.Fatalf("tar member %d = %q, want %q", i, names[i], name)
		}
	}
}

func assertChecksumLine(t *testing.T, checksums, archivePath, basename string) {
	t.Helper()

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(data)
	wantHex := hex.EncodeToString(sum[:])
	wantLine := wantHex + "  " + basename

	if !strings.Contains(checksums, wantLine) {
		t.Fatalf("checksums missing %q", wantLine)
	}
}

func hostReleaseArchive(version string) (string, bool) {
	var goos string

	switch runtime.GOOS {
	case "linux":
		goos = "linux"
	case "darwin":
		goos = "darwin"
	default:
		return "", false
	}

	var goarch string

	switch runtime.GOARCH {
	case "amd64":
		goarch = "amd64"
	case "arm64":
		goarch = "arm64"
	default:
		return "", false
	}

	return fmt.Sprintf("aws-cli-guard_%s_%s_%s.tar.gz", version, goos, goarch), true
}

func releaseAssetsHandler(archive string, archiveBytes []byte, checksums string) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch filepath.Base(request.URL.Path) {
		case archive:
			_, _ = responseWriter.Write(archiveBytes)
		case "checksums.txt":
			_, _ = io.WriteString(responseWriter, checksums)
		default:
			http.NotFound(responseWriter, request)
		}
	})
}

func runInstallScriptWithVersion(
	t *testing.T,
	root, version, prefix, baseURL, realBinDir string,
) ([]byte, error) {
	t.Helper()

	script := filepath.Join(root, "install.sh")
	cmd := exec.CommandContext(
		t.Context(),
		"sh",
		script,
		"--version",
		version,
		"--prefix",
		prefix,
		"--base-url",
		baseURL,
	)

	cmd.Env = installScriptEnv(realBinDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("install.sh: %w", err)
	}

	return output, nil
}

func installScriptEnv(realBinDir string) []string {
	sep := string(os.PathListSeparator)
	path := realBinDir + sep + "/usr/bin" + sep + "/bin"
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
