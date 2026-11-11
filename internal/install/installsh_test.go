//go:build unix

package install_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInstallScriptHappyPath(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	record := filepath.Join(work, "record.txt")

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, record)

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)

	server := httptest.NewServer(releaseTestHandler(archiveBytes, checksums, archiveName))
	defer server.Close()

	output, err := runInstallScript(t, root, server.URL)
	if err != nil {
		t.Fatalf("install.sh: %v\n%s", err, output)
	}

	got, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(got)) != "install" {
		t.Fatalf("stub got %q, want install", got)
	}
}

func TestInstallScriptRejectsBadChecksum(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, filepath.Join(work, "record.txt"))

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	checksums := "0000000000000000000000000000000000000000000000000000000000000000  " +
		archiveName + "\n"

	server := httptest.NewServer(releaseTestHandler(archiveBytes, checksums, archiveName))
	defer server.Close()

	output, err := runInstallScript(t, root, server.URL)
	if err == nil {
		t.Fatal("expected install.sh to fail")
	}

	if !strings.Contains(string(output), "checksum mismatch") {
		t.Fatalf("output = %q, want checksum mismatch", output)
	}
}

func releaseTestHandler(archiveBytes []byte, checksums, archiveName string) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch filepath.Base(request.URL.Path) {
		case archiveName:
			_, _ = responseWriter.Write(archiveBytes)
		case "checksums.txt":
			_, _ = io.WriteString(responseWriter, checksums)
		default:
			http.NotFound(responseWriter, request)
		}
	})
}

func runInstallScript(t *testing.T, root, baseURL string) ([]byte, error) {
	t.Helper()

	script := filepath.Join(root, "install.sh")
	cmd := exec.CommandContext(
		t.Context(),
		"sh",
		script,
		"--version",
		"1.2.3",
		"--base-url",
		baseURL,
	)

	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("install.sh: %w", err)
	}

	return output, nil
}

func writeStubGuard(t *testing.T, path, record string) {
	t.Helper()

	body := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\n", record)

	err := os.WriteFile(path, []byte(body), 0o755)
	if err != nil {
		t.Fatal(err)
	}
}

func buildReleaseArchive(t *testing.T, guardPath string) []byte {
	t.Helper()

	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, name := range []string{archiveMemberAWS, archiveMemberGuard, archiveMemberDisable} {
		var data []byte

		var mode int64 = 0o755

		if name == archiveMemberGuard {
			var err error

			data, err = os.ReadFile(guardPath)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			data = []byte("#!/bin/sh\nexit 0\n")
		}

		hdr := &tar.Header{ //nolint:exhaustruct_v5 // test archive omits unused tar metadata fields.
			Name:     name,
			Mode:     mode,
			Size:     int64(len(data)),
			Typeflag: tar.TypeReg,
			ModTime:  time.Unix(0, 0).UTC(),
		}

		err := tarWriter.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}

		_, err = tarWriter.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := tarWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = gzipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
