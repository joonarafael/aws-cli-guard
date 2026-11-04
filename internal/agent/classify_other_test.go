//go:build !unix

//nolint:testpackage // asserts Observe results through unexported stdin field.
package agent

import (
	"os"
	"testing"
)

func TestObservePipeUndeterminedOnNonUnix(t *testing.T) {
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
	if ctx.stdin != StdinUndetermined {
		t.Fatalf("stdin = %v, want undetermined", ctx.stdin)
	}
}
