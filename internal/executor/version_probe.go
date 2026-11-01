//go:build unix

package executor

import (
	"aws-cli-guard/internal/cliversion"
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

const probeStreamLimit = 64 * 1024

func probeCLIVersion(realPath string) cliversion.Result {
	cmd := exec.CommandContext(context.Background(), realPath, "--version")
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &limitedWriter{dest: &stdout, limit: probeStreamLimit}
	cmd.Stderr = &limitedWriter{dest: &stderr, limit: probeStreamLimit}

	err := cmd.Run()
	_ = err

	return cliversion.ClassifyOutput(stdout.Bytes(), stderr.Bytes())
}

type limitedWriter struct {
	dest  *bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(data []byte) (int, error) {
	remain := w.limit - w.dest.Len()
	if remain <= 0 {
		return len(data), nil
	}

	if len(data) > remain {
		data = data[:remain]
	}

	wrote, writeErr := w.dest.Write(data)
	if writeErr != nil {
		return wrote, fmt.Errorf("aws-cli-guard: version probe buffer: %w", writeErr)
	}

	if len(data) > wrote {
		return len(data), nil
	}

	return wrote, nil
}
