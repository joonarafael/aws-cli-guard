# Development Guide

This document describes how to build, test, and dogfood AWS CLI Guard on a developer machine.

## Prerequisites

- **Go** — same major version as the `go` line in [go.mod](../go.mod). Linux and macOS are supported development hosts.
- **C compiler** — required for `make test` because tests run with the Go race detector (`CGO_ENABLED=1`). `make build` stays a pure Go binary (`CGO_ENABLED=0`).
- **No AWS account required** — integration tests use a fake AWS executable; you do not need credentials to run the test suite.
- **Windows** — not a supported development or runtime target yet. See ADR 0001.

## Day-to-day workflow

From the repository root:

```bash
make help
make check    # Go lint, 300-line file cap, Markdown lint, vet, race-enabled tests
make build    # CGO_ENABLED=0 wrapper in bin/aws
make lint-fix # auto-fix Go formatting when golangci-lint can
make md-fmt   # auto-fix Markdown
make md-lint  # Markdown check only
```

`golangci-lint` is pinned in [go.mod](../go.mod) as a Go `tool` dependency. Nothing needs to be installed globally; Make invokes `go tool golangci-lint`.

To bump the pinned golangci-lint version:

```bash
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@vX.Y.Z
go mod tidy
```

Markdown is checked with [rumdl](https://github.com/rvben/rumdl), pinned by `RUMDL_VERSION` in the [Makefile](../Makefile). The first `make md-lint` or `make md-fmt` downloads that release into `.cache/rumdl` and verifies its SHA-256. Line length is disabled in [.rumdl.toml](../.rumdl.toml). Relative links and heading fragments are checked. `make md-fmt` rewrites fixable issues and still exits successfully when something has to be fixed by hand; `make md-lint` is the check CI runs.

Run `make check` before pushing. GitHub Actions runs the same gate on Linux for pull requests and pushes to `master` (see [.github/workflows/ci.yml](../.github/workflows/ci.yml)).

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

- **Linux** — primary development machine; run `make check` here routinely. CI runs on `ubuntu-24.04`.
- **macOS** — run `make check` on the laptop before merging changes that touch `internal/executor/exec_unix.go`, PATH lookup, or other platform-sensitive code.

There is no Windows CI or manual Windows testing in this phase. `make lint` and `make vet` still typecheck the `//go:build !unix` stub under `GOOS=windows`.

## How integration tests work

Tests live in [internal/executor/executor_test.go](../internal/executor/executor_test.go). They compile the wrapper and [testdata/fakeaws](../internal/executor/testdata/fakeaws), spawn the wrapper as a subprocess, and assert on exit codes and recorded arguments. When the test binary is built with `-race`, those subprocess builds use `-race` as well. `syscall.Exec` does not return in the wrapper process, so the wrapper cannot be exercised entirely in-process.

## Architecture decisions

New significant choices should get an ADR under [docs/decisions/](decisions/). Number them sequentially (`0002-...`, etc.).
