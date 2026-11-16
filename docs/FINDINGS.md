# Findings log

This file records non-obvious discoveries that changed code or rules. Plain design choices belong in [ARCHITECTURE.md](ARCHITECTURE.md), not here.

Entries are append-only. If behavior changes, add a note under the original entry rather than rewriting history.

## F-01: Environment variables could redirect policy and grants

Date: 2026-11-01

Found: Early designs allowed `AWS_CLI_GUARD_*_DIR`, `XDG_*`, and symlinked `HOME/.aws` layouts to relocate guard config, state, and cache.

Why it mattered: A process that controls environment could point the wrapper at attacker-controlled policy while the user believes defaults apply.

What we do now: Paths come from the user database home with `HOME` fallback only. `XDG_*` is ignored. `AWS_CLI_GUARD_*_DIR` compiles only under `awsguard_testpaths`.

Where: [internal/paths](../internal/paths/)

## F-02: CGO-free os/user reads only /etc/passwd

Date: 2026-11-01

Found: Without CGO, `os/user.Current()` does not see LDAP/SSSD nss entries.

Why it mattered: Corporate laptops could resolve a different home from `HOME` than from the user database, shifting guard paths unexpectedly.

What we do now: Accept `HOME` fallback without canceling; document as a bypass in README and [DESIGN.md](DESIGN.md).

Where: [internal/paths/home.go](../internal/paths/home.go)

## F-03: Two wrapper copies on PATH recurse forever

Date: 2026-11-01

Found: Each wrapper skipped only itself on `PATH`, so two builds handed off to each other indefinitely.

Why it mattered: Dogfood plus installed wrapper broke every command.

What we do now: Skip candidates that share file identity or the same Go `buildinfo` main package as the running wrapper. No re-entry env marker — that would break nested `aws` from `credential_process`.

Where: [internal/executor/find.go](../internal/executor/find.go)

## F-04: encoding/json silently keeps the last duplicate key

Date: 2026-10-28

Found: Standard library JSON unmarshaling does not reject duplicate object keys.

Why it mattered: `grants.json` could contain two `profile` keys; the silent winner could hide an attack.

What we do now: Pre-scan JSON objects with a custom duplicate-key detector before decode.

Where: [internal/decode/json_scan.go](../internal/decode/json_scan.go)

## F-05: go-toml Error() can quote file contents

Date: 2026-10-28

Found: Pelletier go-toml error strings may include snippets of the parsed document.

Why it mattered: Policy files can hold sensitive account labels; errors must stay secret-safe.

What we do now: Copy only `Key()` and `Position()` into wrapper errors.

Where: [internal/decode/toml.go](../internal/decode/toml.go)

## F-06: Go panics print argument values

Date: 2026-11-01

Found: Stack traces and panic values include function parameters.

Why it mattered: Passing access keys or tokens into helpers could leak in crash output.

What we do now: Keep secrets out of parameters; redact verbose stacks.

Where: [DESIGN.md](DESIGN.md) secret-handling themes; [internal/output](../internal/output/)

## F-07: Version 1 grant fingerprints ignored credential chains

Date: 2026-11-02

Found: Fingerprints without `source_profile`, `credential_source`, and `sso_session` chains could match the wrong grant.

Why it mattered: A grant for a development hop profile could approve production after config edits.

What we do now: Grants schema v2 requires `chain`; migrate rewrites v1 files.

Where: [internal/grants](../internal/grants/)

## F-08: Unknown global options shift command detection

Date: 2026-11-16

Found: Globals with values added in AWS CLI 2.32–2.40 (`--cli-error-format`, `--output off`, `--help`, `aws login`, `login_session`) change how argv is classified.

Why it mattered: `configure get` and login flows could be misidentified, skipping or misapplying policy.

What we do now: Parser table through ceiling 2.40.3; login/login_session handling tracked for future issues.

Where: [internal/invocation/globals.go](../internal/invocation/globals.go)

## F-09: Darwin terminal check on /dev/null returns ENODEV

Date: 2026-11-05

Found: `TIOCGETA` on `/dev/null` is `ENODEV`, not `ENOTTY`.

Why it mattered: Stdin classification thought `/dev/null` was a TTY.

What we do now: Treat `ENODEV` like non-tty for classification.

Where: [internal/agent/classify_unix.go](../internal/agent/classify_unix.go)

## F-10: macOS stat fields and Fsync durability differ

Date: 2026-11-02

Found: Linux `syscall.Stat_t` fields and `Fsync` semantics differ from Darwin; naive `syscall.Fsync` is not durable on macOS.

Why it mattered: CLI version cache keys and partial files could be wrong or torn after crash.

What we do now: Platform stat shims; `os.File.Sync` for writes (uses `F_FULLFSYNC` on macOS).

Where: [internal/executor/stat_linux.go](../internal/executor/stat_linux.go), [internal/executor/stat_darwin.go](../internal/executor/stat_darwin.go), [internal/files/write_unix.go](../internal/files/write_unix.go)

## F-11: Kernel session IDs are reused across terminals

Date: 2026-11-08

Found: `getsid` alone does not identify a physical terminal session across disconnect/reconnect.

Why it mattered: Session-off records could leak across terminals sharing a session id.

What we do now: Bind `off` records to controlling tty device identity; macOS uses `fcntl(F_GETPATH)`.

Where: [internal/session](../internal/session/), [internal/sessions](../internal/sessions/)

## F-12: CLAUDECODE is set in ordinary IDE terminals

Date: 2026-11-08

Found: `CLAUDECODE` appears in human IDE shells; `CLAUDE_CODE_CHILD_SESSION` is inherited into screen and background launchers.

Why it mattered: False positives would block legitimate human use or miss agents depending on choice.

What we do now: Catalog excludes `CLAUDECODE`; documents child-session marker with pinned vendor version.

Where: [internal/agent/markers_catalog.go](../internal/agent/markers_catalog.go)

## F-13: rename replaces existing install files

Date: 2026-11-14

Found: Using rename without exclusive create could replace an existing wrapper mid-install.

Why it mattered: Partial upgrades could leave an inconsistent prefix.

What we do now: Installer uses `O_CREAT|O_EXCL` for new install files.

Where: [internal/install](../internal/install/)

## F-14: Upgrading could truncate the running binary

Date: 2026-11-14

Found: Replacing `aws` in the install prefix while executing from that file truncates the mapped executable.

Why it mattered: In-place upgrade from the installed wrapper could self-corrupt.

What we do now: Same-file check refuses unsafe replace paths.

Where: [internal/install](../internal/install/), [internal/update](../internal/update/)

## F-15: Wrapper cannot network after exec or export env to shell

Date: 2026-11-14

Found: After `exec`, no wrapper code runs; before `exec`, network stalls the user command. Child `aws-guard off` cannot export state to the parent shell.

Why it mattered: Auto-update and session state need side channels.

What we do now: Detached `Setsid` update helper; per-session cache files; session-off persisted in `sessions.json` only.

Where: [internal/update](../internal/update/), [internal/cache](../internal/cache/)

## F-16: syscall.Exec breaks in-process integration tests

Date: 2026-11-01

Found: The wrapper process disappears on allow, so `go test` cannot assert exec in-process.

Why it mattered: Untagged tests could touch the developer's real policy paths.

What we do now: Subprocess tests with `awsguard_testpaths` and fakeaws; document in [CONTRIBUTING.md](../CONTRIBUTING.md).

Where: [internal/executor/testenv_test.go](../internal/executor/testenv_test.go)

## F-17: govulncheck crashes or moves with Go pin

Date: 2026-11-14

Found: Pinned `govulncheck` against newer toolchains can crash; stdlib findings shift when Go version changes.

Why it mattered: Blocking CI on flaky scanner output hid real regressions.

What we do now: Report findings in job summary and PR comment; job does not fail on vulncheck alone.

Where: [scripts/ci/govulncheck-report.sh](../scripts/ci/govulncheck-report.sh)

## F-18: AWS CLI v1 and v2 tags share one repository

Date: 2026-11-16

Found: `aws/aws-cli` carries both `1.*` and `2.*` tags. Tag lists from `git ls-remote` are not semver-ordered. Patch releases can add global options (for example `2.39.6` added `--help`).

Why it mattered: A naive "latest tag" scan or a comment on every weekly run would miss v2 releases or spam the tracker when only the version string moved.

What we do now: Filter `2.*`, sort numerically, and diff `awscli/data/cli.json` global options on every ahead-of-ceiling release—not only on minor bumps.

Where: [scripts/ci/aws-cli-ceiling-watch.sh](../scripts/ci/aws-cli-ceiling-watch.sh)
