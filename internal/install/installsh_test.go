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

func TestInstallPublicKeyMatchesRepo(t *testing.T) {
	t.Parallel()

	assertInstallKeyMatchesRepo(t, moduleRoot(t))
}

func TestInstallScriptHappyPath(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()
	record := filepath.Join(work, "record.txt")

	secPath, pubB64 := generateMinisignKeypair(t)

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, record)

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)
	sig := signChecksums(t, secPath, []byte(checksums))

	server := httptest.NewServer(releaseTestHandler(archiveBytes, checksums, archiveName, sig))
	defer server.Close()

	output, err := runInstallScript(t, root, server.URL, pubB64)
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

	secPath, pubB64 := generateMinisignKeypair(t)

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, filepath.Join(work, "record.txt"))

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	checksums := "0000000000000000000000000000000000000000000000000000000000000000  " +
		archiveName + "\n"
	sig := signChecksums(t, secPath, []byte(checksums))

	server := httptest.NewServer(releaseTestHandler(archiveBytes, checksums, archiveName, sig))
	defer server.Close()

	output, err := runInstallScript(t, root, server.URL, pubB64)
	if err == nil {
		t.Fatal("expected install.sh to fail")
	}

	if !strings.Contains(string(output), "checksum mismatch") {
		t.Fatalf("output = %q, want checksum mismatch", output)
	}
}

func TestInstallScriptRejectsMissingSignature(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	_, pubB64 := generateMinisignKeypair(t)

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, filepath.Join(work, "record.txt"))

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)

	server := httptest.NewServer(releaseTestHandler(archiveBytes, checksums, archiveName, nil))
	defer server.Close()

	_, err := runInstallScript(t, root, server.URL, pubB64)
	if err == nil {
		t.Fatal("expected install.sh to fail")
	}
}

func TestInstallScriptRejectsBadSignature(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	_, pubB64 := generateMinisignKeypair(t)

	stubGuard := filepath.Join(work, "stub-guard.sh")
	writeStubGuard(t, stubGuard, filepath.Join(work, "record.txt"))

	archiveBytes := buildReleaseArchive(t, stubGuard)
	archiveName := hostArchiveForVersion("1.2.3")
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)

	server := httptest.NewServer(
		releaseTestHandler(archiveBytes, checksums, archiveName, []byte("bad\n")),
	)
	defer server.Close()

	output, err := runInstallScript(t, root, server.URL, pubB64)
	if err == nil {
		t.Fatal("expected install.sh to fail")
	}

	if !strings.Contains(string(output), "checksums signature invalid") {
		t.Fatalf("output = %q, want signature error", output)
	}
}

func releaseTestHandler(
	archiveBytes []byte,
	checksums, archiveName string,
	sig []byte,
) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch filepath.Base(request.URL.Path) {
		case archiveName:
			_, _ = responseWriter.Write(archiveBytes)
		case "checksums.txt":
			_, _ = io.WriteString(responseWriter, checksums)
		case "checksums.txt.minisig":
			if sig == nil {
				http.NotFound(responseWriter, request)

				return
			}

			_, _ = responseWriter.Write(sig)
		default:
			http.NotFound(responseWriter, request)
		}
	})
}

func runInstallScript(t *testing.T, root, baseURL, pubB64 string) ([]byte, error) {
	t.Helper()

	script := writeInstallScriptCopy(t, root, pubB64)
	cmd := exec.CommandContext(
		t.Context(),
		"sh",
		script,
		"--version",
		"1.2.3",
		"--base-url",
		baseURL,
	)

	cmd.Env = append(os.Environ(), "PATH="+pathPrepend(minisignDir(t)))

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
