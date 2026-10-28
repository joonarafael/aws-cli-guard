package terminal_test

import (
	"aws-cli-guard/internal/terminal"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

type fakeTTY struct {
	in  io.Reader
	out *bytes.Buffer
}

func (tty *fakeTTY) Read(p []byte) (int, error) {
	n, err := tty.in.Read(p)
	if err != nil {
		return n, fmt.Errorf("fake tty read: %w", err)
	}

	return n, nil
}

func (tty *fakeTTY) Write(p []byte) (int, error) {
	n, err := tty.out.Write(p)
	if err != nil {
		return n, fmt.Errorf("fake tty write: %w", err)
	}

	return n, nil
}

func TestConfirmOnReadWriterExactMatch(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{
		in:  strings.NewReader("production-admin\n"),
		out: &bytes.Buffer{},
	}

	err := terminal.ConfirmOnReadWriter("production-admin", "type profile name", tty)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}

	if !strings.Contains(tty.out.String(), "aws-cli-guard: type profile name") {
		t.Fatalf("prompt: %q", tty.out.String())
	}
}

func TestConfirmOnReadWriterTrailingNewlineOnly(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{
		in:  strings.NewReader("ok\n"),
		out: &bytes.Buffer{},
	}

	err := terminal.ConfirmOnReadWriter("ok", "prompt", tty)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
}

func TestConfirmOnReadWriterRejectsExtraSpace(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{
		in:  strings.NewReader("ok \n"),
		out: &bytes.Buffer{},
	}

	err := terminal.ConfirmOnReadWriter("ok", "prompt", tty)
	if !errors.Is(err, terminal.ErrMismatch) {
		t.Fatalf("err = %v, want ErrMismatch", err)
	}
}

func TestConfirmOnReadWriterMismatch(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{
		in:  strings.NewReader("wrong\n"),
		out: &bytes.Buffer{},
	}

	err := terminal.ConfirmOnReadWriter("expected", "prompt", tty)
	if !errors.Is(err, terminal.ErrMismatch) {
		t.Fatalf("err = %v, want ErrMismatch", err)
	}

	if strings.Contains(err.Error(), "wrong") {
		t.Fatalf("error must not echo typed input: %v", err)
	}
}

func TestConfirmOnReadWriterEmptyPhrase(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{in: strings.NewReader(""), out: &bytes.Buffer{}}

	err := terminal.ConfirmOnReadWriter("", "prompt", tty)
	if !errors.Is(err, terminal.ErrInvalidPhrase) {
		t.Fatalf("err = %v, want ErrInvalidPhrase", err)
	}
}

func TestConfirmOnReadWriterPhraseWithControl(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{in: strings.NewReader(""), out: &bytes.Buffer{}}

	err := terminal.ConfirmOnReadWriter("bad\nphrase", "prompt", tty)
	if !errors.Is(err, terminal.ErrInvalidPhrase) {
		t.Fatalf("err = %v, want ErrInvalidPhrase", err)
	}
}

func TestConfirmOnReadWriterOverlongLine(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 1025)
	tty := &fakeTTY{
		in:  strings.NewReader(long + "\n"),
		out: &bytes.Buffer{},
	}

	err := terminal.ConfirmOnReadWriter(long, "prompt", tty)
	if !errors.Is(err, terminal.ErrMismatch) {
		t.Fatalf("err = %v, want ErrMismatch", err)
	}
}

func TestConfirmOnReadWriterEOF(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{in: strings.NewReader(""), out: &bytes.Buffer{}}

	err := terminal.ConfirmOnReadWriter("phrase", "prompt", tty)
	if !errors.Is(err, terminal.ErrMismatch) {
		t.Fatalf("err = %v, want ErrMismatch", err)
	}
}

func TestConfirmInvalidPhrase(t *testing.T) {
	t.Parallel()

	tty := &fakeTTY{in: strings.NewReader(""), out: &bytes.Buffer{}}

	err := terminal.ConfirmOnReadWriter("valid", "prompt", tty)
	if !errors.Is(err, terminal.ErrMismatch) {
		t.Fatalf("EOF without input: err = %v", err)
	}

	err = terminal.ConfirmOnReadWriter("", "prompt", tty)
	if !errors.Is(err, terminal.ErrInvalidPhrase) {
		t.Fatal("empty phrase")
	}

	err = terminal.ConfirmOnReadWriter("a\x00b", "prompt", tty)
	if !errors.Is(err, terminal.ErrInvalidPhrase) {
		t.Fatal("control in phrase")
	}
}
