//go:build unix

package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

const (
	readChunkSize = 32 * 1024
	filePermMask  = 0o777
)

// ReadRegular reads a regular file without following a symlink at the final path component.
// On success it returns the file permission bits masked to 0777.
func ReadRegular(path string) ([]byte, os.FileMode, error) {
	fileFD, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, 0, readOpenError(path, err)
	}

	defer func() {
		_ = syscall.Close(fileFD)
	}()

	mode, regErr := assertRegularFD(fileFD, path)
	if regErr != nil {
		return nil, 0, regErr
	}

	data, err := readFromFD(fileFD)
	if err != nil {
		return nil, 0, fmt.Errorf("aws-cli-guard: read %q: %w", path, err)
	}

	return data, mode, nil
}

func readOpenError(path string, err error) error {
	if errors.Is(err, syscall.ELOOP) {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrSymlink, path)
	}

	if errors.Is(err, syscall.ENOENT) {
		return fmt.Errorf("aws-cli-guard: open %q: %w", path, os.ErrNotExist)
	}

	if errors.Is(err, syscall.EISDIR) {
		return fmt.Errorf("aws-cli-guard: %w: %q", ErrNotRegular, path)
	}

	return fmt.Errorf("aws-cli-guard: open %q: %w", path, err)
}

func assertRegularFD(fileFD int, path string) (os.FileMode, error) {
	var stat syscall.Stat_t

	err := syscall.Fstat(fileFD, &stat)
	if err != nil {
		return 0, fmt.Errorf("aws-cli-guard: fstat %q: %w", path, err)
	}

	if stat.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return 0, fmt.Errorf("aws-cli-guard: %w: %q", ErrNotRegular, path)
	}

	return os.FileMode(stat.Mode & filePermMask), nil
}

func readFromFD(fileFD int) ([]byte, error) {
	var out []byte

	buf := make([]byte, readChunkSize)

	for {
		count, err := syscall.Read(fileFD, buf)
		if count > 0 {
			out = append(out, buf[:count]...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return out, nil
			}

			return nil, fmt.Errorf("syscall read: %w", err)
		}

		if count == 0 {
			return out, nil
		}
	}
}
