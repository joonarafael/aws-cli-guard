//go:build unix

//nolint:testpackage // exercises FetchReleaseArchive with fixture signatures.
package update

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
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func buildTestReleaseArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, name := range []string{"aws", "aws-guard", "aws-cli-guard-disable"} {
		data := []byte("#!/bin/sh\nexit 0\n")
		hdr := &tar.Header{ //nolint:exhaustruct_v5 // test archive.
			Name:     name,
			Mode:     0o755,
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

func TestFetchReleaseArchiveRequiresValidSignature(t *testing.T) {
	t.Parallel()

	version := "9.8.7"
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if !SupportedPlatform(goos, goarch) {
		t.Skip("unsupported platform")
	}

	pub, secPath := generateMinisignKeypair(t)
	archiveBytes := buildTestReleaseArchive(t)
	archive := archiveName(version, goos, goarch)
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archive)
	sig := signChecksumsCLI(t, secPath, []byte(checksums))

	server := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch filepath.Base(request.URL.Path) {
			case archive:
				_, _ = responseWriter.Write(archiveBytes)
			case "checksums.txt":
				_, _ = io.WriteString(responseWriter, checksums)
			case "checksums.txt.minisig":
				_, _ = responseWriter.Write(sig)
			default:
				http.NotFound(responseWriter, request)
			}
		}),
	)
	defer server.Close()

	dest := t.TempDir()

	client := ReleaseClient{ //nolint:exhaustruct_v5 // test client; LatestURL unused.
		DownloadBase: server.URL,
		HTTP:         server.Client(),
	}

	err := FetchReleaseArchive(
		client,
		nil,
		version,
		goos,
		goarch,
		dest,
		pub,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"aws", "aws-guard", "aws-cli-guard-disable"} {
		_, statErr := os.Stat(filepath.Join(dest, name))
		if statErr != nil {
			t.Fatalf("%s: %v", name, statErr)
		}
	}
}

func TestFetchReleaseArchiveRejectsBadSignature(t *testing.T) {
	t.Parallel()

	version := "9.8.7"
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if !SupportedPlatform(goos, goarch) {
		t.Skip("unsupported platform")
	}

	pub, _ := generateMinisignKeypair(t)
	archiveBytes := buildTestReleaseArchive(t)
	archive := archiveName(version, goos, goarch)
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archive)

	server := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch filepath.Base(request.URL.Path) {
			case archive:
				_, _ = responseWriter.Write(archiveBytes)
			case "checksums.txt":
				_, _ = io.WriteString(responseWriter, checksums)
			case "checksums.txt.minisig":
				_, _ = io.WriteString(responseWriter, "not a signature\n")
			default:
				http.NotFound(responseWriter, request)
			}
		}),
	)
	defer server.Close()

	client := ReleaseClient{ //nolint:exhaustruct_v5 // test client; LatestURL unused.
		DownloadBase: server.URL,
		HTTP:         server.Client(),
	}

	err := FetchReleaseArchive(
		client,
		nil,
		version,
		goos,
		goarch,
		t.TempDir(),
		pub,
	)
	if err == nil {
		t.Fatal("expected failure")
	}
}
