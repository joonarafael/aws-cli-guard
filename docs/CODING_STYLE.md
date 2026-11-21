# Coding style

Conventions for AWS CLI Guard. The linter configuration enforces most Go rules; this document explains the habits behind it. Day-to-day commands are in [CONTRIBUTING.md](../CONTRIBUTING.md).

## Toolchain and formatting

- Go version matches the `go` line in [go.mod](../go.mod).
- **golangci-lint** runs with `default: all`; each disabled linter has a comment in [.golangci.yml](../.golangci.yml).
- Formatting: **gofumpt**, **goimports**, **golines** at **100** columns (sync with `lll`).

Run `make lint-fix` after Go edits; `make lint` is the check.

## Package and file layout

- One package per concern under `internal/`, with a **doc.go** stating constraints.
- **300 lines maximum per `.go` file**, tests included, enforced by [scripts/lint/go-file-length.sh](../scripts/lint/go-file-length.sh).
- Split large tests into `_cases_test.go`, `_helpers_test.go`, `fixtures_test.go`, `_assert_test.go`.

Entrypoints stay thin in `cmd/`; logic belongs in `internal/`.

## Code conventions

### Environment slices

Pass `[]string` env into libraries. Only `cmd/` calls `os.Environ()`.

### Errors

Use typed, secret-safe errors (`Kind`, `Key`, `Path`, line/column). Never include secret **values**. Wrap with `%w` when context helps.

### Output and cancellations

All wrapper text goes through [internal/output/output.go](../internal/output/output.go) (`Cancel`, `Error`, `Warning`, `Detail`). Cancellations exit **3**.

### Fail closed

Strict decoders reject unknown fields and duplicate JSON keys. Unsupported schema versions are errors, not warnings.

### Schema changes

Bump the file's `version` field, teach `aws-guard migrate`, and document in [ARCHITECTURE.md](ARCHITECTURE.md). Add a [FINDINGS.md](FINDINGS.md) entry when behavior surprises.

### Guard-owned files

Read and write only through [internal/files](../internal/files/) helpers so modes, symlinks, and atomicity stay consistent.

## Platform files and build tags

| Suffix / tag | Use |
| --- | --- |
| `_unix.go` | Linux and Darwin |
| `_linux.go`, `_darwin.go` | OS-specific |
| `_other.go` with `//go:build !unix` | Windows stubs |
| `awsguard_testpaths` | Test path overrides in `internal/paths` |
| `!awsguard_testpaths` | Production ignores `AWS_CLI_GUARD_*_DIR` |

## Tests

- External `_test` packages (`package foo_test`).
- `t.Parallel()` in table tests unless a shared resource forbids it.
- Table cases as functions returning slices.
- Integration tests use **fakeaws** and subprocess wrapper builds.
- `make test` enables the race detector (`CGO_ENABLED=1`).

Assert exit **3** and `aws-cli-guard: error:` on cancellations.

## Dependencies

**depguard** allows only the Go standard library, `aws-cli-guard`, `github.com/pelletier/go-toml/v2`, and `aead.dev/minisign` (minisign signature verification for release checksums). New dependencies need justification in the pull request.

## Lint suppressions

Use `//nolint:lintername // reason` with the linter name and a short reason. Prefer fixing the code when practical.

**exhaustruct_v5** expects struct literals to list every field (with documented ignores for `syscall.SysProcAttr`).

## Comments

Comments state **constraints**, not history. Do not cite removed ADR numbers, roadmap IDs, or GitHub issue numbers in code.

## Commands, shell, and docs

- New `aws-guard` subcommands must appear in [internal/cli/help.go](../internal/cli/help.go) and its coverage test.
- Shell scripts: `#!/bin/sh`, `set -eu`, ShellCheck `-S warning` with project [.shellcheckrc](../.shellcheckrc).
- Markdown: `make md-fmt` / `make md-lint` (rumdl). Prose is plain English in short declarative sentences. Valid relative links only.
