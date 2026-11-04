//go:build unix

//nolint:testpackage // asserts Observe results through unexported stdin field.
package agent

import (
	"os"
	"testing"
)

func TestObservePipeNonInteractive(t *testing.T) {
	t.Parallel()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})

	ctx := Observe(reader)
	if ctx.stdin != StdinNonInteractive {
		t.Fatalf("stdin = %v, want non-interactive", ctx.stdin)
	}
}

func TestObserveDevNullNonInteractive(t *testing.T) {
	t.Parallel()

	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = file.Close() })

	ctx := Observe(file)
	if ctx.stdin != StdinNonInteractive {
		t.Fatalf("stdin = %v, want non-interactive", ctx.stdin)
	}
}

func TestObserveClosedUndetermined(t *testing.T) {
	t.Parallel()

	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx := Observe(file)
	if ctx.stdin != StdinUndetermined {
		t.Fatalf("stdin = %v, want undetermined", ctx.stdin)
	}
}
