# AWS CLI Guard: AI Agents Guide

## What This Is

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2. It adds user-controlled authorization guardrails to AWS profile usage. It is small, understandable, and conservative when it cannot tell whether an invocation is permitted.

Treat it as a local safety mechanism against accidental misuse. Stronger guarantees need controls outside the wrapper. The wrapper does not evaluate IAM, and a caller can still reach the real AWS CLI by an absolute path or another `PATH`.

When a check is ambiguous, stop the command, explain why on stderr, and exit 3. That status is reserved for a cancellation. The real AWS CLI keeps its own exit codes.

| Topic | Document |
| --- | --- |
| Intent | [docs/DESIGN.md](docs/DESIGN.md) |
| Current behavior | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) |
| Lessons learned | [docs/FINDINGS.md](docs/FINDINGS.md) |
| Conventions | [docs/CODING_STYLE.md](docs/CODING_STYLE.md) |
| Workflow | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Planned work | GitHub issues |

**Never read or expose secrets.** That includes `.env` files, AWS credentials, SSO tokens, and any other material that would reveal keys or session data. If you need to know something about the user's environment, ask them.

The default branch is `master`. Pull requests merge into that branch.

You can talk with the user in any language they prefer. The codebase, documentation, commit messages, and pull requests are written in English.

## Where Things Live

| Area | Location | Contents |
| --- | --- | --- |
| Entrypoints | `cmd/aws`, `cmd/aws-guard` | Wrapper and control-plane `main` |
| Libraries | `internal/` | Packages that exist today |
| Docs | `docs/` | Design, architecture, findings, coding style |
| Contributing | `CONTRIBUTING.md` | Build, test, maintain, contribute |
| GitHub | `.github/` | PR template, issue forms, workflows |
| Lint scripts | `scripts/lint/` | Go file length, rumdl, golangci-lint config check |
| Build | `Makefile` | Day-to-day targets |
| CI | `.github/workflows/ci.yml` | Hygiene gate, plus secret scan, workflow lint, ShellCheck, govulncheck |

How to build, test, and dogfood is [CONTRIBUTING.md](CONTRIBUTING.md). Read it before changing the workflow.

Keep this map short. Update it when a new top-level area appears. Leave function-level detail in the code.

## Tooling

Formatting and linting are **strict and automated**. CI rejects non-compliant code. **Always run the appropriate formatters and linters on files you touch before finishing.**

| What | Fix | Check |
| --- | --- | --- |
| Go | `make lint-fix` | `make lint` |
| Go file length (300 lines, tests included) | Split the file | `make file-length` |
| Markdown | `make md-fmt` | `make md-lint` |
| Vet and tests |  | `make vet`, `make test` |
| Local gate |  | `make check` |

`golangci-lint` is pinned in [go.mod](go.mod) as a Go `tool` dependency. Make runs `go tool golangci-lint`. Nothing has to be installed globally. To bump the pin, see [CONTRIBUTING.md](CONTRIBUTING.md).

Markdown uses [rumdl](https://github.com/rvben/rumdl), pinned by `RUMDL_VERSION` in the [Makefile](Makefile). The first `make md-lint` or `make md-fmt` downloads that release into `.cache/rumdl` and checks its SHA-256. `make md-fmt` rewrites what it can and still exits successfully when a problem has to be fixed by hand. `make md-lint` is the check CI runs.

Shell is POSIX `sh` only (`#!/bin/sh`, `set -eu`) and must pass ShellCheck. CI runs `shellcheck -S warning scripts/lint/*.sh`. Dialect and optional checks are in [.shellcheckrc](.shellcheckrc).

Go is the version on the `go` line in `go.mod`. Linux and macOS are the supported development hosts. `make test` needs a C compiler because tests run with the race detector (`CGO_ENABLED=1`). `make build` stays a pure Go binary (`CGO_ENABLED=0`). Tests use a fake AWS executable, so they need no AWS account and no credentials.

Style is [docs/CODING_STYLE.md](docs/CODING_STYLE.md) plus [.golangci.yml](.golangci.yml). Match the code around your change. Fix the code when a linter fails.

## Debugging

**Debug problems yourself.** Trace the failure through to a cause.

1. Read the output. A cancellation exits with status **3**. The two cancellation lines start with `aws-cli-guard: error:`; further guidance (diagnosis, allow hints) uses `aws-cli-guard:` detail lines. A `default` profile warning may appear on stderr before a cancellation. A warning that leaves the command running starts with `aws-cli-guard: warning:`. After an allowed command, the process is the real AWS CLI and the exit code is the CLI's.
2. Reproduce with `make test`, or with `go test` in the package you changed.
3. For a behavior question, read [docs/DESIGN.md](docs/DESIGN.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) before editing.
4. Confirm `go version` matches `go.mod`. A sandbox may have different tools than the user's machine.

`make test` and `make check` are safe to run. They do not call AWS.

Dogfood by prepending `bin/` to `PATH` for the current shell, as [CONTRIBUTING.md](CONTRIBUTING.md) describes. Leave `~/.local/bin` alone until an installer exists.

## Linters and Formatters

The project enforces strict lint and format rules. Follow them.

Run the linters and formatters while you work. Fix each failure as it appears, before they pile up.

## Before You Finish

1. Run formatters and linters again on every file you changed.
2. If you touched Go, run `make lint-fix` and `make file-length`.
3. If you touched shell, run `shellcheck -S warning` on those scripts.
4. If you touched documentation, run `make md-fmt`, then `make md-lint`.
5. If you changed code, run `make test`.
6. Run `make check` before you call the work done. That is the local gate: Go lint, the 300-line file cap, Markdown lint, vet, and race-enabled tests. CI runs that gate, then the build, gitleaks, actionlint, and ShellCheck. govulncheck still runs in that job. A finding or a scanner crash is written to the job summary and a sticky pull-request comment, and does not fail the job.
7. Keep the change small, and match the patterns already in the surrounding code.
8. Update [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) when behavior or a contract changes; add [docs/FINDINGS.md](docs/FINDINGS.md) for a non-obvious pitfall; put future work in a GitHub issue, not in docs.

If you added files, keep each directory small and split by concern. The size limit this repository enforces is 300 lines per Go file, tests included.

## Git

While you work, notice when the diff against `origin/master` is getting large. Recommend splitting it into smaller pull requests, and help with the split when the user wants that. Large mechanical refactors, code moves, and tests are the usual exceptions.

Write pull requests in English.
