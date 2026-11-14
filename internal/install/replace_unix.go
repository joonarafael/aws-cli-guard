//go:build unix

package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"time"
)

const maxReplacePartialAttempts = 100

//nolint:unparam // callers pass modeExecutable; the parameter keeps the write path explicit.
func replaceRegularFile(dst, src string, mode os.FileMode) error {
	if sameInstallPath(dst, src) {
		return nil
	}

	err := refuseSymlink(dst)
	if err != nil {
		return err
	}

	//nolint:gosec // G304: src is a validated install sibling.
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: open %q: %w", src, err)
	}

	defer func() {
		_ = srcFile.Close()
	}()

	now := time.Now()

	var lastErr error

	for i := range maxReplacePartialAttempts {
		partial := partialReplacePath(dst, now, i)

		writeErr := writePartialFromReader(partial, srcFile, mode)
		if writeErr != nil {
			if partialReplaceRetryable(writeErr) {
				lastErr = writeErr

				_, _ = srcFile.Seek(0, io.SeekStart)

				continue
			}

			return writeErr
		}

		err = refuseSymlink(dst)
		if err != nil {
			_ = os.Remove(partial)

			return err
		}

		err = os.Rename(partial, dst)
		if err != nil {
			_ = os.Remove(partial)

			return fmt.Errorf("aws-cli-guard: rename %q to %q: %w", partial, dst, err)
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return ErrReplacePartial
}

func replaceManifestFile(dst string, data []byte) error {
	err := refuseSymlink(dst)
	if err != nil {
		return err
	}

	now := time.Now()

	var lastErr error

	for i := range maxReplacePartialAttempts {
		partial := partialReplacePath(dst, now, i)

		writeErr := writePartialBytes(partial, data, modeManifest)
		if writeErr != nil {
			if partialReplaceRetryable(writeErr) {
				lastErr = writeErr

				continue
			}

			return writeErr
		}

		err = refuseSymlink(dst)
		if err != nil {
			_ = os.Remove(partial)

			return err
		}

		err = os.Rename(partial, dst)
		if err != nil {
			_ = os.Remove(partial)

			return fmt.Errorf("aws-cli-guard: rename %q to %q: %w", partial, dst, err)
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return ErrReplacePartial
}

func partialReplacePath(livePath string, now time.Time, attempt int) string {
	stamp := now.UTC().Format("20060102T150405.000000000Z")

	base := livePath + ".partial." + stamp
	if attempt == 0 {
		return base
	}

	return base + "." + strconv.Itoa(attempt+1)
}

func partialReplaceRetryable(err error) bool {
	return errors.Is(err, syscall.EEXIST) || errors.Is(err, syscall.ELOOP)
}

func writePartialFromReader(partial string, srcFile *os.File, mode os.FileMode) error {
	partialFD, err := syscall.Open(
		partial,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW,
		uint32(mode),
	)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: create partial %q: %w", partial, err)
	}

	dstFile := os.NewFile(uintptr(partialFD), partial)

	copyErr := copyAndClose(dstFile, partial, srcFile)
	if copyErr != nil {
		return copyErr
	}

	chmodErr := os.Chmod(partial, mode)
	if chmodErr != nil {
		_ = os.Remove(partial)

		return fmt.Errorf("aws-cli-guard: chmod partial %q: %w", partial, chmodErr)
	}

	return nil
}

func writePartialBytes(partial string, data []byte, mode os.FileMode) error {
	partialFD, err := syscall.Open(
		partial,
		syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY|syscall.O_NOFOLLOW,
		uint32(mode),
	)
	if err != nil {
		return fmt.Errorf("aws-cli-guard: create partial %q: %w", partial, err)
	}

	file := os.NewFile(uintptr(partialFD), partial)

	if len(data) > 0 {
		_, err = file.Write(data)
		if err != nil {
			_ = file.Close()
			_ = os.Remove(partial)

			return fmt.Errorf("aws-cli-guard: write partial %q: %w", partial, err)
		}
	}

	err = file.Sync()
	if err != nil {
		_ = file.Close()
		_ = os.Remove(partial)

		return fmt.Errorf("aws-cli-guard: sync partial %q: %w", partial, err)
	}

	err = file.Close()
	if err != nil {
		_ = os.Remove(partial)

		return fmt.Errorf("aws-cli-guard: close partial %q: %w", partial, err)
	}

	chmodErr := os.Chmod(partial, mode)
	if chmodErr != nil {
		_ = os.Remove(partial)

		return fmt.Errorf("aws-cli-guard: chmod partial %q: %w", partial, chmodErr)
	}

	return nil
}
