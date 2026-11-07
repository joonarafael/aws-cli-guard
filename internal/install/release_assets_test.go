//go:build unix

package install_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const releaseTestVersion = "1.2.3"

func TestReleaseAssetsRejectsBadVersions(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)

	for _, bad := range []string{"v1.2.3", "1.2", "1.2.3-rc"} {
		t.Run(bad, func(t *testing.T) {
			t.Parallel()

			outdir := filepath.Join(t.TempDir(), "out")

			err := runReleaseAssets(t, root, bad, outdir)
			if err == nil {
				t.Fatal("expected release-assets.sh to fail")
			}

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
				t.Fatalf("err = %v, want exit 3", err)
			}

			_, statErr := os.Stat(outdir)
			if statErr == nil {
				entries, readErr := os.ReadDir(outdir)
				if readErr != nil {
					t.Fatal(readErr)
				}

				if len(entries) != 0 {
					t.Fatalf("outdir has %d entries, want none", len(entries))
				}
			}
		})
	}
}

func TestReleaseAssetsArchivesAndChecksums(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	outdir := t.TempDir()

	err := runReleaseAssets(t, root, releaseTestVersion, outdir)
	if err != nil {
		t.Fatalf("release-assets.sh: %v", err)
	}

	wantArchives := []string{
		fmt.Sprintf("aws-cli-guard_%s_linux_amd64.tar.gz", releaseTestVersion),
		fmt.Sprintf("aws-cli-guard_%s_linux_arm64.tar.gz", releaseTestVersion),
		fmt.Sprintf("aws-cli-guard_%s_darwin_amd64.tar.gz", releaseTestVersion),
		fmt.Sprintf("aws-cli-guard_%s_darwin_arm64.tar.gz", releaseTestVersion),
	}

	for _, name := range wantArchives {
		path := filepath.Join(outdir, name)

		_, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("missing %s: %v", name, statErr)
		}

		assertTarMembers(t, path)
	}

	repoInstall, err := os.ReadFile(filepath.Join(root, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}

	distInstall, err := os.ReadFile(filepath.Join(outdir, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(repoInstall, distInstall) {
		t.Fatal("dist install.sh differs from repository install.sh")
	}

	checksumsPath := filepath.Join(outdir, "checksums.txt")

	checksumsBytes, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatal(err)
	}

	checksumsText := string(checksumsBytes)
	for _, name := range wantArchives {
		assertChecksumLine(t, checksumsText, filepath.Join(outdir, name), name)
	}
}

func TestReleaseAssetsInstallScriptWithBuiltArchive(t *testing.T) {
	t.Parallel()

	archive, ok := hostReleaseArchive(releaseTestVersion)
	if !ok {
		t.Skip("unsupported test host for release archive")
	}

	root := moduleRoot(t)
	work := t.TempDir()
	outdir := filepath.Join(work, "dist")
	prefix := filepath.Join(work, "prefix")

	err := runReleaseAssets(t, root, releaseTestVersion, outdir)
	if err != nil {
		t.Fatalf("release-assets.sh: %v", err)
	}

	archivePath := filepath.Join(outdir, archive)

	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archive)

	server := httptest.NewServer(releaseAssetsHandler(archive, archiveBytes, checksums))
	defer server.Close()

	realDir := filepath.Join(work, "real-bin")
	buildFakeAWS(t, root, realDir)

	output, err := runInstallScriptWithVersion(
		t,
		root,
		releaseTestVersion,
		prefix,
		server.URL,
		realDir,
	)
	if err != nil {
		t.Fatalf("install.sh: %v\n%s", err, output)
	}

	manifest, err := os.ReadFile(filepath.Join(prefix, "aws-cli-guard-manifest"))
	if err != nil {
		t.Fatal(err)
	}

	wantLine := "version=" + releaseTestVersion
	if !strings.Contains(string(manifest), wantLine) {
		t.Fatalf("manifest = %q, want %q", manifest, wantLine)
	}
}
