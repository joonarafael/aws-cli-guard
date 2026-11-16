# ADR 0033: aws-guard help and version

## Status

Accepted

## Context

First-time users install from a release and need to discover control-plane commands without reading the repository. Today bare `aws-guard`, unknown commands, and `--help` all stop with a single usage error line and exit 3, which reads like a failure. Release builds stamp `internal/version.Version`; there was no supported way to print it.

Per-command `--help` (for example `aws-guard allow --help`) is useful later but adds parser work in every subcommand. A top-level catalog is enough for v0.1.0.

## Decision

1. **Commands.** `aws-guard help`, `aws-guard --help`, and `aws-guard -h` print an ordered command list to **stdout** and exit **0**. `aws-guard version` and `aws-guard --version` print `aws-guard <Version>` to stdout ( `dev` when unstamped ) and exit **0**.

2. **Catalog.** One table in `internal/cli/help.go` lists every top-level `Dispatch` command with a short description and a synopsis derived from each command’s existing `usage:` string (the `usage: aws-guard` prefix stripped). Rows for `init`, `help`, and `version` use fixed synopses.

3. **Usage failures.** With no subcommand or an unknown subcommand, stderr still starts with `aws-cli-guard: error: usage: aws-guard <command>`, then prints the same catalog on stderr, and exits **3** per [ADR 0003](0003-wrapper-exit-status.md). Argv must not be echoed.

4. **Out of scope.** Subcommand-specific `--help` stays undefined for now; those flags follow each command’s existing usage handling.

5. **Tests.** A drift guard fails CI when `Dispatch` handles a command that is missing from the help table.

## Consequences

- [DEVELOPMENT.md](../DEVELOPMENT.md) next free ADR number is `0034`.
- Help text must be updated when a new top-level command is added to `Dispatch`.
- `aws-guard -h` is only the control-plane help flag, not the `aws` wrapper.
