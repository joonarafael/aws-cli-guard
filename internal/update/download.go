package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxChecksumsSize = 1 << 20
	maxArchiveSize   = 64 << 20
	sha256HexLen     = 64
)

// SupportedPlatform reports whether auto-update may download for goos/goarch.
func SupportedPlatform(goos, goarch string) bool {
	switch goos {
	case "linux", "darwin":
		switch goarch {
		case "amd64", "arm64":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// FetchReleaseArchive downloads and verifies the platform archive into destDir.
func FetchReleaseArchive(
	client ReleaseClient,
	httpClient *http.Client,
	version, goos, goarch, destDir string,
) error {
	if httpClient == nil {
		httpClient = client.HTTP
	}

	if httpClient == nil {
		//nolint:exhaustruct_v5 // default HTTP client.
		httpClient = &http.Client{Timeout: requestTimeout}
	}

	base := releaseDownloadBase(client, version)
	archive := archiveName(version, goos, goarch)

	err := downloadFile(
		httpClient,
		base+"/"+archive,
		filepath.Join(destDir, archive),
		maxArchiveSize,
	)
	if err != nil {
		return err
	}

	err = downloadFile(
		httpClient,
		base+"/checksums.txt",
		filepath.Join(destDir, "checksums.txt"),
		maxChecksumsSize,
	)
	if err != nil {
		return err
	}

	return verifyArchive(filepath.Join(destDir, archive), filepath.Join(destDir, "checksums.txt"))
}

//nolint:wrapcheck // stdlib HTTP and file errors stay diagnostic at call sites.
func downloadFile(client *http.Client, url, dest string, maxSize int64) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errDownloadHTTP, resp.StatusCode)
	}

	//nolint:gosec // G304: dest comes from a verified release download path.
	file, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	_, err = io.Copy(file, io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return err
	}

	info, err := os.Stat(dest)
	if err != nil {
		return err
	}

	if info.Size() > maxSize {
		return errDownloadTooLarge
	}

	return nil
}

//nolint:wrapcheck // read and unpack errors are returned to helper outcomes.
func verifyArchive(archivePath, checksumsPath string) error {
	//nolint:gosec // paths come from a verified release download directory.
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		return err
	}

	if len(data) > maxChecksumsSize {
		return errChecksumsTooLarge
	}

	base := filepath.Base(archivePath)

	line, found := findChecksumLine(string(data), base)
	if !found {
		return errChecksumLineMissing
	}

	expected, ok := parseChecksumLine(line, base)
	if !ok {
		return errChecksumLineInvalid
	}

	//nolint:gosec // archivePath is under a temp download directory.
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		return err
	}

	sum := sha256.Sum256(archiveData)

	actual := hex.EncodeToString(sum[:])
	if actual != expected {
		return errChecksumMismatch
	}

	return unpackArchive(archivePath, filepath.Dir(archivePath))
}

func findChecksumLine(body, archive string) (string, bool) {
	for line := range strings.SplitSeq(body, "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "  "+archive) {
			return strings.TrimSpace(line), true
		}
	}

	return "", false
}

func parseChecksumLine(line, archive string) (string, bool) {
	parts := strings.Split(line, "  ")
	if len(parts) != 2 || parts[1] != archive {
		return "", false
	}

	if len(parts[0]) != sha256HexLen {
		return "", false
	}

	_, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", false
	}

	return parts[0], true
}

//nolint:funlen,gosec,wrapcheck // install.sh archive rules; members are validated.
func unpackArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	defer func() {
		_ = gzipReader.Close()
	}()

	tarReader := tar.NewReader(gzipReader)

	seen := make(map[string]struct{})

	for {
		header, readErr := tarReader.Next()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return readErr
		}

		name := header.Name

		err := validateArchiveMember(name)
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			return errArchiveMemberRegular
		}

		seen[name] = struct{}{}

		dest := filepath.Join(destDir, name)

		out, createErr := os.Create(dest)
		if createErr != nil {
			return createErr
		}

		_, copyErr := io.Copy(out, tarReader)
		closeErr := out.Close()

		if copyErr != nil {
			return copyErr
		}

		if closeErr != nil {
			return closeErr
		}
	}

	want := []string{"aws", "aws-guard", "aws-cli-guard-disable"}
	if len(seen) != len(want) {
		return errArchiveMemberCount
	}

	for _, name := range want {
		if _, ok := seen[name]; !ok {
			return errArchiveMemberMissing
		}
	}

	return nil
}

func validateArchiveMember(name string) error {
	switch {
	case name == "aws", name == "aws-guard", name == "aws-cli-guard-disable":
		return nil
	case strings.HasPrefix(name, "/"):
		return errArchiveAbsolutePath
	case strings.Contains(name, "/"):
		return errArchiveDirectory
	case name == ".." || strings.HasPrefix(name, ".."):
		return errArchiveParentPath
	default:
		return errArchiveUnexpected
	}
}
