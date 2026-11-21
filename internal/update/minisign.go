package update

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aead.dev/minisign"
)

const (
	maxSignatureSize        = 8 << 10
	checksumsTrustedFile    = "file:checksums.txt"
	checksumsSignatureAsset = "checksums.txt.minisig"
)

//go:embed minisign.pub
var embeddedPublicKeyFile []byte

// EmbeddedPublicKey is the release signing key committed in the repository.
func EmbeddedPublicKey() (minisign.PublicKey, error) {
	var key minisign.PublicKey

	err := key.UnmarshalText(embeddedPublicKeyFile)
	if err != nil {
		return minisign.PublicKey{}, fmt.Errorf("%w: %w", errSignatureInvalid, err)
	}

	return key, nil
}

func verifyChecksumsSignature(publicKey minisign.PublicKey, checksums, signature []byte) error {
	if len(signature) == 0 {
		return errSignatureMissing
	}

	if len(signature) > maxSignatureSize {
		return errSignatureTooLarge
	}

	var sig minisign.Signature

	err := sig.UnmarshalText(signature)
	if err != nil {
		return errSignatureInvalid
	}

	if sig.KeyID != publicKey.ID() {
		return errSignatureInvalid
	}

	if !trustedCommentNamesChecksums(sig.TrustedComment) {
		return errSignatureUntrusted
	}

	rawSignature, err := sig.MarshalText()
	if err != nil {
		return errSignatureInvalid
	}

	ok := verifyChecksumsBytes(publicKey, checksums, rawSignature, sig.Algorithm)
	if !ok {
		return errSignatureUntrusted
	}

	return nil
}

func trustedCommentNamesChecksums(comment string) bool {
	for field := range strings.FieldsSeq(comment) {
		if field == checksumsTrustedFile {
			return true
		}
	}

	return false
}

func downloadAndVerifyChecksums(
	httpClient *http.Client,
	base, destDir string,
	publicKey minisign.PublicKey,
) error {
	err := downloadFile(
		httpClient,
		base+"/checksums.txt",
		filepath.Join(destDir, "checksums.txt"),
		maxChecksumsSize,
	)
	if err != nil {
		return err
	}

	sigPath := filepath.Join(destDir, "checksums.txt.minisig")

	err = downloadFile(httpClient, base+"/"+checksumsSignatureAsset, sigPath, maxSignatureSize)
	if err != nil {
		return err
	}

	checksumsPath := filepath.Join(destDir, "checksums.txt")

	//nolint:gosec // sigPath is under a verified release download directory.
	sigData, err := os.ReadFile(sigPath)
	if err != nil {
		return err //nolint:wrapcheck // read errors stay diagnostic.
	}

	//nolint:gosec // checksumsPath is under a verified release download directory.
	checksumsData, err := os.ReadFile(checksumsPath)
	if err != nil {
		return err //nolint:wrapcheck // read errors stay diagnostic.
	}

	if len(checksumsData) > maxChecksumsSize {
		return errChecksumsTooLarge
	}

	return verifyChecksumsSignature(publicKey, checksumsData, sigData)
}

func verifyChecksumsBytes(
	publicKey minisign.PublicKey,
	checksums, rawSignature []byte,
	algorithm uint16,
) bool {
	switch algorithm {
	case minisign.HashEdDSA:
		reader := minisign.NewReader(bytes.NewReader(checksums))
		_, err := io.Copy(io.Discard, reader)

		return err == nil && reader.Verify(publicKey, rawSignature)
	case minisign.EdDSA:
		return minisign.Verify(publicKey, checksums, rawSignature)
	default:
		return false
	}
}
