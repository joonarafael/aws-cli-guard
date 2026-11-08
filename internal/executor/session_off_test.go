//go:build linux

package executor_test

import (
	"aws-cli-guard/internal/session"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestWrapperSessionOffHandoff(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, denyProfilePolicy())

	slavePath, dev, ino, rdev := openPTYDeviceLinux(t)
	writeSessionOffFile(t, work, session.ID(), slavePath, dev, ino, rdev)

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"sts",
		"get-caller-identity",
		"--profile",
		enforceProfile,
		cancellationSecretArg,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)
	cmd.Stdin = nil

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Run: %v stderr=%q", err, stderr.String())
	}

	out := stderr.String()
	if !strings.Contains(out, session.SessionOffWarning) {
		t.Fatalf("stderr = %q", out)
	}

	if strings.Contains(out, "deadbeef") {
		t.Fatal("stderr must not contain token")
	}

	_, statErr := os.Stat(record)
	if statErr != nil {
		t.Fatalf("fake AWS must run: %v", statErr)
	}
}

func TestWrapperSessionOffMissingDeviceStillDenied(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	work := t.TempDir()

	writeEnforcePolicy(t, work, denyProfilePolicy())
	writeSessionOffFile(t, work, session.ID(), "/no/such/tty", 1, 2, 3)

	wrapper := build(t, root, filepath.Join(work, "aws"), "./cmd/aws")
	fake := build(t, root, filepath.Join(work, "fake-aws"), "./internal/executor/testdata/fakeaws")
	record := filepath.Join(work, "record.json")

	cmd := exec.CommandContext(
		t.Context(),
		wrapper,
		"sts",
		"get-caller-identity",
		"--profile",
		enforceProfile,
	)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Env = enforceEnv(t, work, fake, record)

	err := cmd.Run()
	assertExitCode(t, err)

	_, statErr := os.Stat(record)
	if statErr == nil {
		t.Fatal("fake AWS must not run")
	}
}

func openPTYDeviceLinux(t *testing.T) (string, uint64, uint64, uint64) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = master.Close() })

	var unlock int32

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCSPTLCK,
		uintptr(unsafe.Pointer(&unlock)),
	)
	if errno != 0 {
		t.Fatal(errno)
	}

	var ptyNum int32

	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCGPTN,
		uintptr(unsafe.Pointer(&ptyNum)),
	)
	if errno != 0 {
		t.Fatal(errno)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", ptyNum)

	info, err := os.Stat(slavePath)
	if err != nil {
		t.Fatal(err)
	}

	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("no stat")
	}

	return slavePath, sys.Dev, sys.Ino, sys.Rdev
}

func writeSessionOffFile(t *testing.T, work, sessionID, ttyPath string, dev, ino, rdev uint64) {
	t.Helper()

	stateDir := filepath.Join(work, "guard-state")

	err := os.MkdirAll(stateDir, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	body := fmt.Sprintf(`{
  "version": 1,
  "records": [
    {
      "session_id": %q,
      "effect": "off",
      "token": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "at": "2026-01-01T00:00:00Z",
      "tty": {
        "path": %q,
        "dev": %d,
        "ino": %d,
        "rdev": %d
      }
    }
  ]
}
`, sessionID, ttyPath, dev, ino, rdev)

	err = os.WriteFile(filepath.Join(stateDir, "sessions.json"), []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
