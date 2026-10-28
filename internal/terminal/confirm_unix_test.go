//go:build unix

package terminal_test

import (
	"aws-cli-guard/internal/terminal"
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const (
	childEnvFlag   = "AWS_CLI_GUARD_TERMINAL_CONFIRM_CHILD"
	childPhraseEnv = "AWS_CLI_GUARD_TERMINAL_CONFIRM_PHRASE"
	exitNoTerminal = 10
)

func TestMain(m *testing.M) {
	if os.Getenv(childEnvFlag) == "1" {
		phrase := os.Getenv(childPhraseEnv)

		err := terminal.Confirm(phrase, "type the phrase")
		if errors.Is(err, terminal.ErrNoTerminal) {
			os.Exit(exitNoTerminal)
		}

		if err == nil {
			os.Exit(0)
		}

		os.Exit(1)
	}

	os.Exit(m.Run())
}

//nolint:paralleltest // Setsid changes process session; must not run parallel with other tests.
func TestConfirmRejectsStdinPipeInNewSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^$")

	cmd.Env = append(os.Environ(),
		childEnvFlag+"=1",
		childPhraseEnv+"=secret-phrase",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	var stderr bytesCapture

	cmd.Stderr = &stderr

	//nolint:exhaustruct_v5 // only Setsid is required for this test.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	_, err = stdin.Write([]byte("secret-phrase\n"))
	if err != nil {
		t.Fatal(err)
	}

	_ = stdin.Close()

	err = cmd.Wait()

	if ctx.Err() != nil {
		t.Fatal("child did not exit")
	}

	exitErr := &exec.ExitError{}

	ok := errors.As(err, &exitErr)
	if !ok {
		t.Fatalf("wait: %v", err)
	}

	if exitErr.ExitCode() != exitNoTerminal {
		t.Fatalf(
			"exit = %d, want %d (stdin pipe must not satisfy confirmation)",
			exitErr.ExitCode(),
			exitNoTerminal,
		)
	}
}

type bytesCapture []byte

func (capture *bytesCapture) Write(p []byte) (int, error) {
	*capture = append(*capture, p...)

	return len(p), nil
}
