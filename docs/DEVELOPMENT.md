# Development Guide

This document describes how to build, test, and dogfood AWS CLI Guard on a developer machine.

## Prerequisites

- **Go** — same major version as the `go` line in [go.mod](../go.mod). Linux and macOS are supported development hosts.
- **No C compiler** — the current code uses only the Go standard library.
- **No AWS account required** — integration tests use a fake AWS executable; you do not need credentials to run the test suite.
- **Windows** — not a supported development or runtime target yet. See ADR 0001.

## Day-to-day workflow

From the repository root:

```bash
go test ./...
go build -o bin/aws ./cmd/aws
```

Run tests after every change. Build when you want to try the wrapper in a shell.

### Dogfooding without installing

Do not copy the binary into `~/.local/bin` until an installer exists. Prepend the build output to `PATH` for the current shell session:

```bash
export PATH="$PWD/bin:$PATH"
command -v aws    # should print .../aws-cli-guard/bin/aws
aws --version   # should run the real AWS CLI if it appears later on PATH
```

The wrapper walks `PATH`, skips its own executable, and execs the next `aws` it finds.

### When the real CLI is not on PATH

Point at the real binary explicitly:

```bash
export AWS_CLI_GUARD_REAL_AWS=/usr/local/bin/aws
export PATH="$PWD/bin:$PATH"
aws sts get-caller-identity
```

Integration tests always set `AWS_CLI_GUARD_REAL_AWS` so discovery is deterministic.

## Platform checks

- **Linux** — primary development machine; run `go test ./...` here routinely.
- **macOS** — run `go test ./...` on the laptop before merging changes that touch `internal/executor/exec_unix.go`, PATH lookup, or other platform-sensitive code.

There is no Windows CI or manual Windows testing in this phase.

## How integration tests work

Tests live in [internal/executor/executor_test.go](../internal/executor/executor_test.go). They compile the wrapper and [testdata/fakeaws](../internal/executor/testdata/fakeaws), spawn the wrapper as a subprocess, and assert on exit codes and recorded arguments. `syscall.Exec` does not return in the wrapper process, so the wrapper cannot be exercised entirely in-process.

## Architecture decisions

New significant choices should get an ADR under [docs/decisions/](decisions/). Number them sequentially (`0002-...`, etc.).
