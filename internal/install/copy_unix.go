//go:build unix

package install

import (
	"fmt"
	"io"
	"os"
	"slices"
	"syscall"
)

const (
	modeExecutable = 0o755
	modeManifest   = 0o644
)

func copyFileExclusive(dst, src string, mode os.FileMode) error {
	//nolint:gosec // G304: src is an install sibling path validated before copy.
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: open %q: %w", src, err)
	}

	defer func() {
		_ = srcFile.Close()
	}()

	dstFD, err := syscall.Open(
		dst,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW,
		uint32(mode),
	)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: create %q: %w", dst, err)
	}

	dstFile := os.NewFile(uintptr(dstFD), dst)

	copyErr := copyAndClose(dstFile, dst, srcFile)
	if copyErr != nil {
		_ = os.Remove(dst)

		return copyErr
	}

	chmodErr := os.Chmod(dst, mode)
	if chmodErr != nil {
		_ = os.Remove(dst)

		return fmt.Errorf("aws-cli-guard: chmod %q: %w", dst, chmodErr)
	}

	return nil
}

func writeBytesExclusive(dst string, data []byte, mode os.FileMode) error {
	dstFD, err := syscall.Open(
		dst,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW,
		uint32(mode),
	)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: create %q: %w", dst, err)
	}

	file := os.NewFile(uintptr(dstFD), dst)

	_, writeErr := file.Write(data)
	if writeErr != nil {
		_ = file.Close()
		_ = os.Remove(dst)

		return fmt.Errorf("aws-cli-guard: write %q: %w", dst, writeErr)
	}

	syncErr := file.Sync()
	closeErr := file.Close()

	if syncErr != nil {
		_ = os.Remove(dst)

		return fmt.Errorf("aws-cli-guard: sync %q: %w", dst, syncErr)
	}

	if closeErr != nil {
		_ = os.Remove(dst)

		return fmt.Errorf("aws-cli-guard: close %q: %w", dst, closeErr)
	}

	chmodErr := os.Chmod(dst, mode)
	if chmodErr != nil {
		_ = os.Remove(dst)

		return fmt.Errorf("aws-cli-guard: chmod %q: %w", dst, chmodErr)
	}

	return nil
}

func copyAndClose(dstFile *os.File, dstPath string, srcFile *os.File) error {
	_, err := io.Copy(dstFile, srcFile)
	if err != nil {
		_ = dstFile.Close()

		return fmt.Errorf("aws-cli-guard: copy to %q: %w", dstPath, err)
	}

	err = dstFile.Sync()
	if err != nil {
		_ = dstFile.Close()

		return fmt.Errorf("aws-cli-guard: sync %q: %w", dstPath, err)
	}

	err = dstFile.Close()
	if err != nil {
		return fmt.Errorf("aws-cli-guard: close %q: %w", dstPath, err)
	}

	return nil
}

func removeCreated(paths []string) {
	for _, path := range slices.Backward(paths) {
		_ = os.Remove(path)
	}
}
