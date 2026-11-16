# Contributing

How to build, test, maintain, and contribute to AWS CLI Guard.

## Develop

### Prerequisites

- **Go** — same major version as the `go` line in [go.mod](go.mod). Linux and macOS are supported development hosts.
- **C compiler** — required for `make test` (race detector, `CGO_ENABLED=1`). `make build` uses `CGO_ENABLED=0`.
- **No AWS account** — tests use a fake AWS executable.

Windows is not a supported development or runtime target; `make vet` and `make lint` still type-check `!unix` stubs.

### Day-to-day workflow

```bash
make help
make check    # Go lint, file cap, Markdown lint, vet, race-enabled tests
make build    # bin/aws and bin/aws-guard
make lint-fix
make md-fmt
make md-lint
```

`golangci-lint` is pinned in [go.mod](go.mod) as a `go tool` dependency.

Markdown uses [rumdl](https://github.com/rvben/rumdl), pinned by `RUMDL_VERSION` in the [Makefile](Makefile). The first `make md-lint` or `make md-fmt` downloads it into `.cache/rumdl` with SHA-256 verification.

Run `make check` before pushing. CI runs the same gate on Linux ([.github/workflows/ci.yml](.github/workflows/ci.yml)). Release tags run `make check` before assets ([.github/workflows/release.yml](.github/workflows/release.yml)).

### Dogfooding without installing

```bash
export PATH="$PWD/bin:$PATH"
command -v aws
aws-guard init
aws --version
```

`make build` also stages `bin/aws-cli-guard-disable` for local install testing.

### Installing from a local build

```bash
make build
aws-guard install
# or: aws-guard install --prefix /path/to/dir
```

Default prefix: `~/.local/share/aws-cli-guard`. A second install refuses while files remain. Use `aws-guard shell bash` or `aws-guard shell zsh` for persistent `PATH`, then `aws-guard init` when needed.

Release assets locally:

```bash
make dist VERSION=X.Y.Z
```

Push tag `vX.Y.Z` to create a **draft** GitHub Release; publish manually after verification. Delete a failed draft before re-tagging. Bump `VERSION=` in [README.md](README.md) Quick start when publishing.

Every guarded command, including `aws --version`, goes through policy. The wrapper checks CLI **2.9.0** floor and **2.40.3** verified ceiling ([docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)).

Scratch directories without touching live files:

```bash
CGO_ENABLED=0 go build -tags=awsguard_testpaths -o bin/aws ./cmd/aws
CGO_ENABLED=0 go build -tags=awsguard_testpaths -o bin/aws-guard ./cmd/aws-guard
export AWS_CLI_GUARD_CONFIG_DIR=/path/to/scratch/config
export AWS_CLI_GUARD_STATE_DIR=/path/to/scratch/state
export AWS_CLI_GUARD_CACHE_DIR=/path/to/scratch/cache
export PATH="$PWD/bin:$PATH"
aws-guard init
```

### When the real CLI is not on PATH

```bash
export AWS_CLI_GUARD_REAL_AWS=/usr/local/bin/aws
export PATH="$PWD/bin:$PATH"
```

Integration tests always set `AWS_CLI_GUARD_REAL_AWS`.

### Tests and build tags

`make test` runs with `-tags=awsguard_testpaths`, then runs `go test ./internal/paths` and `./internal/agent` **without** the tag.

Do not run `go test ./...` without the tag: subprocess tests could read your real home directory.

### Platform checks

- **Linux** — primary CI host (`ubuntu-24.04`).
- **macOS** — run `make check` on a laptop before merging changes to `internal/paths`, `exec_unix.go`, PATH discovery, tty capture, or file durability.

### Integration tests

Subprocess tests in [internal/executor](internal/executor/) compile fakeaws and `./cmd/aws`. Cancellations exit **3** with `aws-cli-guard: error:` on the first lines.

### CI scope

Linux job: `make check`, build, gitleaks, actionlint, ShellCheck, govulncheck report. **govulncheck does not fail the job** — findings land in the job summary and a sticky PR comment ([FINDINGS.md](docs/FINDINGS.md) F-17).

Verbose troubleshooting: `AWS_CLI_GUARD_LOG=verbose` or `level = "verbose"` in `settings.toml`.

## Maintain

### Bump golangci-lint

```bash
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@vX.Y.Z
go mod tidy
```

Align `GOLANGCI_LINT_VERSION` in the [Makefile](Makefile) and [scripts/lint/golangci-config.sh](scripts/lint/golangci-config.sh).

### Bump rumdl

Update `RUMDL_VERSION` in the [Makefile](Makefile).

### Bump Go

Update `go.mod`, `GO_VERSION` in [ci.yml](.github/workflows/ci.yml) and [release.yml](.github/workflows/release.yml), and `run.go` in [.golangci.yml](.golangci.yml).

### Raise the AWS CLI verified ceiling

1. Read AWS CLI changelog since the current ceiling.
2. Check parity with `awscli/data/cli.json` if available.
3. Update [internal/invocation/globals.go](internal/invocation/globals.go) and constants in [internal/cliversion/version.go](internal/cliversion/version.go).
4. Add tests.
5. Update ceiling in [README.md](README.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
6. Add a [FINDINGS.md](docs/FINDINGS.md) entry when behavior shifts.

### Cut a release

1. Push tag `vX.Y.Z`.
2. Verify the draft release assets.
3. Publish manually on GitHub.
4. Delete failed drafts before re-tagging.
5. Bump README Quick start `VERSION=`.

## Contribute

- Open a GitHub **issue** before large features; planned work lives in issues, not in docs.
- Never paste credentials or AWS config in issues or PRs.
- Keep pull requests small; target **master**.
- Fill in [.github/pull_request_template.md](.github/pull_request_template.md).
- Behavior and contracts → [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Intent changes → [docs/DESIGN.md](docs/DESIGN.md). Pitfalls → [docs/FINDINGS.md](docs/FINDINGS.md).
- Follow [docs/CODING_STYLE.md](docs/CODING_STYLE.md) and report security issues per [SECURITY.md](SECURITY.md).

Working agreements for AI agents are in [AGENTS.md](AGENTS.md).

## Project history

The original ADRs, roadmap, and long-form design document are preserved at git tag **`adr-archive`** on commit `a032c80`.
