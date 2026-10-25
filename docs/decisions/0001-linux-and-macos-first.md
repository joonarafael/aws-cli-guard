# ADR 0001: Linux and macOS first

## Status

Accepted

## Context

AWS CLI Guard is a local wrapper around AWS CLI v2. It must run on users' machines, integrate with PATH, replace the current process with the real AWS CLI on success, and eventually inspect parent processes and terminals in an OS-specific way.

The project aims to be cross-platform in the long term. Go makes portable CLI tooling straightforward. The maintainers have a Linux desktop and a macOS laptop for day-to-day development; there is no Windows machine available for implementation or manual testing.

Open question 1 in [DESIGN.md](../DESIGN.md) asked which operating systems to support initially.

## Decision

1. **Cross-platform intent.** Shared code uses portable APIs from the standard library (`os`, `path/filepath`, `os.PathListSeparator`). Avoid hardcoded path separators or Unix-only assumptions outside clearly marked build tags.

2. **Initial supported targets.** Linux and macOS, on both amd64 and arm64. These are the only platforms we implement, test, and document for the first releases.

3. **Development workflow.** Primary development and CI-style checks run on Linux. Before calling platform-sensitive changes done, run `go test ./...` on macOS as well (especially changes to process replacement or PATH discovery).

4. **Windows.** Not implemented in the initial phase. Code may compile on other GOOS values via stub files (`//go:build !unix`) that return a clear error, but we do not claim Windows support, add Windows CI, or optimize for Windows UX until there is a way to test it.

5. **Process replacement on Unix.** On `linux` and `darwin`, approved invocations use `syscall.Exec` to replace the wrapper process with the real AWS CLI. This preserves the terminal, signal handling, and exit status. Non-Unix builds use a separate file that fails with an explicit "not supported yet" message rather than silently misbehaving.

## Consequences

- Users on Windows must wait for a later phase or use WSL without official support until Windows is explicitly added.
- Platform-specific features (parent process inspection, session identifiers) will be added behind interfaces with Linux and macOS implementations first.
- A future ADR can extend support to native Windows when testing infrastructure exists.
