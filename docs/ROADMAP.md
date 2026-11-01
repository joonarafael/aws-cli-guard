# Roadmap and Epic Tracker

A high-level overview of the epics needed to build AWS CLI Guard as described in [DESIGN.md](DESIGN.md). It tracks progress and shows what can be picked up next. Technical detail stays in the design document and in the code.

> **Subject to change.** The ordering and dependencies below are a best current guess, not a plan of record. This is not a waterfall. Implementation will surface surprises, and epics will be split, merged, reordered, or dropped as we learn. Update this file when that happens.

## How to read this

- **ID** is an area prefix and a number, for example `PO-3`. IDs are stable. A dropped epic keeps its ID and is marked dropped rather than renumbered.
- **Status** is one of `Done`, `In progress`, `To do`, or `Dropped`.
- **Milestone** maps to the phases in the [Initial Roadmap](DESIGN.md#15-initial-roadmap). `MVP` is Phase 1. `Backlog` holds Phase 6 and the [recorded operational criteria](DESIGN.md#12-recorded-operational-criteria), which are intentionally not scheduled yet.
- **Blocked by** lists what has to be done, or decided, first. **Blocks** is the reverse link. Keep both sides in sync when editing.
- An epic is ready to start when everything in its **Blocked by** column is `Done` or answered.
- `Q-n` items are design decisions. An open one blocks epics just like another epic does. An answered one points at an ADR and no longer blocks.

| Prefix | Area |
| --- | --- |
| `Q` | Open design decisions |
| `FD` | Foundation |
| `PR` | Profile resolution |
| `FS` | File storage and recovery |
| `PO` | Policy |
| `GR` | Grants |
| `MV` | MVP milestone |
| `SS` | SSO |
| `CI` | Continuous integration |
| `AG` | Agent awareness |
| `IN` | Installation |
| `OP` | Operational criteria |
| `AD` | Advanced features |

## Next up

The MVP is complete. Phase 2 starts with one fix to shipped behavior, then the rest of the milestone:

- `FD-6` skipping every copy of the wrapper during real AWS CLI discovery, so two copies cannot hand off to each other forever.

Then the rest of Phase 2:

- `FD-7` checking the real AWS CLI version against the 2.9.0 floor and verified ceiling, with per-session caching ([ADR 0002](decisions/0002-aws-cli-version-floor.md)).
- `FS-6` loading and validating `settings.toml`, plus `aws-guard recover settings` ([ADR 0006](decisions/0006-identity-verification-off-by-default.md)).
- `GR-7` revoking grants (`aws-guard revoke <profile>` and `aws-guard revoke --all`).
- `PR-5` detecting credential providers beyond the environment-key case and bumping `grants.json` to version 2, followed by `PR-6` showing account and role in messages.
- `PO-6` adding a dedicated policy key for environment credentials, followed by `PO-5` inspect, explain, and validate policy.
- `SS-1` login handoff ([ADR 0004](decisions/0004-login-out-of-scope.md)).

Also unblocked, but in later milestones: `GR-6`, `AG-1`, `IN-1`, `OP-3`, `OP-5`, `AD-2`.

`Q-1` through `Q-9` are answered. None of them block current work.

`CI-1` is unblocked too. Leave it until `AG-4`. The note on `CI-1` says why.

## Suggested order for the MVP

Each wave only depends on earlier waves. Items within a wave can run in parallel.

1. `GR-1`. Done. `MV-2`, `FS-1`, `FD-3`, `FS-2`, `FS-3`, `FD-4`, and `PR-4` were already done.
2. `PO-2`. Done. `PR-3` and `GR-2` were already done.
3. `PO-3`, `FS-4`, and `GR-3`. Done.
4. `PO-4`, `GR-4`. Done.
5. `MV-1`. Done.

After the MVP, the design suggests Phase 2 (profile awareness), Phase 3 (grant operations), Phase 4 (agent awareness), and Phase 5 (installation), in that order. The design also says later phases should follow actual user needs, so that order is a default rather than a commitment.

## Milestones

| Milestone | Epics |
| --- | --- |
| MVP (Phase 1) | `FD-1`, `FD-2`, `FD-3`, `FD-4`, `PR-1`, `PR-2`, `PR-3`, `PR-4`, `FS-1`, `FS-2`, `FS-3`, `FS-4`, `PO-1`, `PO-2`, `PO-3`, `PO-4`, `GR-1`, `GR-2`, `GR-3`, `GR-4`, `MV-1`, `MV-2` |
| Phase 2: Better profile awareness | `FD-5`, `FD-6`, `FD-7`, `PR-5`, `PR-6`, `FS-6`, `PO-5`, `PO-6`, `GR-7`, `SS-1` |
| Phase 3: Grant operations | `GR-5`, `GR-6` |
| Phase 4: AI-agent awareness | `FS-5`, `AG-1`, `AG-2`, `AG-3`, `CI-1`, `AG-4`, `AG-5` |
| Phase 5: Installation and portability | `IN-1`, `IN-2`, `IN-3` |
| Backlog | `OP-1`, `OP-2`, `OP-3`, `OP-4`, `OP-5`, `AD-1`, `AD-2`, `AD-3`, `AD-4`, `AD-5` |

The MVP is complete when every MVP epic is `Done`.

## Open design decisions

Most of these come from [Open Design Questions](DESIGN.md#16-open-design-questions). The rest are decisions the design leaves open in other sections.

| ID | Decision | Status | Blocks |
| --- | --- | --- | --- |
| Q-1 | Which shells get first-class integration? | Answered in [ADR 0005](decisions/0005-bash-and-zsh.md) | — |
| Q-2 | Is `aws sso login` allowed by default, or does it get its own policy? | Answered in [ADR 0004](decisions/0004-login-out-of-scope.md) | — |
| Q-3 | Is identity verification optional, and when does it run? | Answered in [ADR 0006](decisions/0006-identity-verification-off-by-default.md) | — |
| Q-4 | Which AI agents are supported initially? | Answered in [ADR 0015](decisions/0015-initial-agent-integrations.md) | — |
| Q-5 | Is agent detection on by default, or opt-in? | Answered in [ADR 0016](decisions/0016-agent-detection-switches.md) | — |
| Q-6 | Which minimum AWS CLI version and configuration features are supported? | Answered in [ADR 0002](decisions/0002-aws-cli-version-floor.md) | — |
| Q-7 | Which mechanism does persistent disable and enable use? | Answered in [ADR 0017](decisions/0017-persistent-disable-swaps-aws.md) | — |
| Q-8 | Is auto-update on or off by default? | Answered in [ADR 0007](decisions/0007-auto-update-asked-at-init.md) | — |
| Q-9 | Which exit status does the wrapper use for its own refusals? | Answered in [ADR 0003](decisions/0003-wrapper-exit-status.md) | — |

Q-1 to Q-6 are design questions 2, 4, 5, 8, 9, and 12. Q-7 comes from [Persistent system-wide disable and enable](DESIGN.md#122-persistent-system-wide-disable-and-enable), Q-8 from [Auto-update and update notices](DESIGN.md#124-auto-update-and-update-notices), and Q-9 from the exit-code requirement in [User Experience](DESIGN.md#9-user-experience). All of them are answered.

## Epics

### FD: Foundation

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| FD-1 | Wrapper entrypoint and real AWS CLI handoff | Done | MVP | — | `FD-6`, `FD-7` |
| FD-2 | On-disk path resolution | Done | MVP | — | `FS-1`, `FD-5` |
| FD-3 | Wrapper exit status and secret-safe error messages | Done | MVP | — | `PR-3`, `FS-3`, `OP-5`, `AD-2` |
| FD-4 | `aws-guard` command skeleton and terminal confirmation | Done | MVP | — | `FS-4`, `PO-2`, `GR-3`, `GR-7`, `OP-1` |
| FD-5 | Guard-owned paths from the user database | Done | Phase 2 | `FD-2` | — |
| FD-6 | Skip every wrapper copy during real AWS CLI discovery | To do | Phase 2 | `FD-1` | `FD-7` |
| FD-7 | AWS CLI version check and session-cached warning | To do | Phase 2 | `FD-1`, `FD-6`, `FS-1`, `FS-2` | — |

- `FD-5` implements [ADR 0014](decisions/0014-guard-paths-from-user-database.md) in `internal/paths`:
  - The home directory comes from `os/user.Current()`. `HOME` is the fallback when the lookup fails or returns an empty or non-absolute directory. The fallback is supported, prints no warning, and does not cancel.
  - `XDG_*` variables are no longer read.
  - The `AWS_CLI_GUARD_*_DIR` overrides compile only under the `awsguard_testpaths` build tag. `make test` and the integration-test build helper pass that tag; `make build` does not.
  - Put only the `AWS_CLI_GUARD_*_DIR` environment lookup behind two small build-tagged files (`override_testpaths.go` with `//go:build awsguard_testpaths` and `override_prod.go` with `//go:build !awsguard_testpaths`), while keeping the `os/user.Current()` lookup and `HOME` fallback in shared, untagged code in `internal/paths/paths.go`. That way both `make build` and `make test` compile and lint the real user-database resolution logic, and `internal/executor` integration tests only need `-tags=awsguard_testpaths` when compiling the `./cmd/aws` subprocess binary.
  - The lookup is injectable for unit tests.
  - Run `make check` on macOS before merging. Update [DEVELOPMENT.md](DEVELOPMENT.md) with the tag and the new dogfooding behavior.
- `FD-6` closes two self-handoff loops in `internal/executor/find.go`. Today `AWS_CLI_GUARD_REAL_AWS` is returned without a self-check, and `PATH` discovery only skips the running file. With two wrapper copies on `PATH`, each finds the other first. The design is in [Core Architecture](DESIGN.md#3-core-architecture):
  - A candidate is a wrapper when it is the same file as the running executable, or when `debug/buildinfo.ReadFile` reports the same main package path that `debug.ReadBuildInfo` reports for the running wrapper. A file with no Go build information is not a wrapper.
  - `PATH` discovery skips wrappers. An `AWS_CLI_GUARD_REAL_AWS` that names a wrapper cancels with status 3, without printing the path and without falling back to `PATH`.
  - In tests, `./cmd/aws` has `Main.Path == "aws-cli-guard/cmd/aws"` while `fakeaws` has `Main.Path == "aws-cli-guard/internal/executor/testdata/fakeaws"`, so comparing `Main.Path` naturally distinguishes the wrapper from `fakeaws`.
  - Tests: a second wrapper copy earlier on `PATH` is skipped and the fake AWS runs once; the override naming the wrapper, or a copy of it, exits 3 and the fake does not run; the fake AWS, which is a different Go main package, is not treated as a wrapper.
  - Roughly 30 lines of code plus tests. It implements part of the Phase 5 safe-detection item early, the same way `PR-4` was pulled ahead.
- `FD-7` implements [ADR 0002](decisions/0002-aws-cli-version-floor.md):
  - Checks the resolved real AWS CLI's version against the `2.9.0` floor and the verified ceiling constant.
  - Older than `2.9.0`, or an unreadable version, logs a `aws-cli-guard: warning:` line and hands off to the real CLI with no policy evaluation. Newer than the verified ceiling logs a warning and continues into the normal guarded path.
  - Caches the detected version and whether the current terminal session (`getsid(0)` on Linux and macOS, with parent-PID fallback when `getsid` is unavailable) has already been warned in a versioned file under `~/.cache/aws-cli-guard/`, keyed by the resolved binary's metadata (`path`, `device`, `inode`, `size`, `mtime`). Subsequent invocations in the same session reuse the cached version without spawning `<real-aws> --version` and emit the warning at most once per session. Cache misses or corrupt cache files do not block `aws`.
  - Updates `internal/executor/testdata/fakeaws` to report a supported version on `--version` by default, with an environment override for testing `<2.9.0`, `>ceiling`, unreadable output, and per-session caching.

### PR: Profile resolution

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| PR-1 | Credential subject classification | Done | MVP | `PR-2` | `PR-3`, `PO-3`, `PO-6` |
| PR-2 | AWS CLI global option parsing | Done | MVP | — | `PR-1`, `PO-4`, `SS-1` |
| PR-3 | Warning when the `default` profile is used | Done | MVP | `PR-1`, `FD-3` | `MV-1` |
| PR-4 | Reading AWS config and credentials files, including custom locations | Done | MVP | — | `PR-5`, `GR-2` |
| PR-5 | Detecting credential providers: SSO, role chaining, `credential_process` | To do | Phase 2 | `PR-4`, `GR-1`, `GR-2` | `PR-6`, `FS-5`, `GR-5`, `AD-1` |
| PR-6 | Showing account and role in messages | To do | Phase 2 | `PR-5` | — |

- `PR-3` warns on stderr when the credential subject is the named profile `default`. The command still follows policy for `default`.
- `PR-4` is pulled ahead of Phase 2. The design lists custom AWS config locations under Phase 2, but the grant fingerprint in `GR-2` is MVP work and needs to read the AWS config file correctly.
- `PR-5` classifies credential providers and resolves `[sso-session]`, `source_profile`, `credential_source`, and role chains (with cycle detection) in `internal/awsconfig` to power `PR-6` refusal details and `PO-5` `policy explain`. It bumps `grants.json` to `version = 2` (while still supporting `version = 1` in memory) so `grants.Fingerprint` and `grants.Source` bind role chains, `credential_source`, and `sso-session` settings, as anticipated in [ADR 0011](decisions/0011-grants-schema.md) and [ADR 0012](decisions/0012-credential-source-fingerprint.md).
- `PR-6` uses the provider resolution from `PR-5` to include grounded account and role detail lines in refusal messages when available in config.

### FS: File storage and recovery

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| FS-1 | Safe file handling: atomic writes, permissions, symlink refusal, locking | Done | MVP | `FD-2` | `FD-7`, `FS-4`, `FS-5`, `FS-6`, `PO-1`, `GR-1`, `OP-1` |
| FS-2 | Strict, versioned file decoding | Done | MVP | — | `FD-7`, `FS-3`, `FS-5`, `FS-6`, `PO-1`, `GR-1` |
| FS-3 | Diagnosis of corrupted files | Done | MVP | `FS-2`, `FD-3` | `FS-4`, `FS-6`, `MV-1` |
| FS-4 | `aws-guard recover` | Done | MVP | `FS-1`, `FS-3`, `FD-4` | `FS-6` |
| FS-5 | Migrate command for older file versions | To do | Phase 4 | `FS-1`, `FS-2`, `PR-5`, `PO-6`, `AG-3` | `IN-3` |
| FS-6 | Preference settings file | To do | Phase 2 | `FS-1`, `FS-2`, `FS-3`, `FS-4` | `AD-1`, `OP-4` |

- `FS-6` is `settings.toml` from [ADR 0006](decisions/0006-identity-verification-off-by-default.md), pulled into Phase 2 because `aws-guard init` (`PO-2`) already creates `settings.toml` on disk. Identity verification and the auto-update answer live there. `FS-6` adds `settings.Load`, strict validation (rejecting `verify = true` until `AD-1` enforces it), corrupt-file diagnosis on the `aws` path (`settings.Corruption` and `diagnose.Write` in `internal/gate`), and `aws-guard recover settings`.
- `FS-5` is placed in Phase 4 so `aws-guard migrate` has real schema transitions to rewrite after `PR-5` (`grants.json` v2), `PO-6` (`config.toml` v2), and `AG-3` (`[agents]` rules) land.
- `FS-3` adds `policy.Corruption` and `diagnose.Write` for policy loads. The wrapper prints the note after `output.Cancel` when a load is corrupt. `FS-4` adds `aws-guard recover policy`, `recover grants`, and `recover cache`. `aws-guard recover settings` lands with `FS-6`.

### PO: Policy

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| PO-1 | Policy schema and loading | Done | MVP | `FS-1`, `FS-2` | `PO-2`, `PO-3`, `PO-6` |
| PO-2 | `aws-guard init` and the missing-policy flow | Done | MVP | `PO-1`, `FD-4` | `MV-1` |
| PO-3 | Policy evaluation engine | Done | MVP | `PO-1`, `PR-1` | `PO-4`, `PO-5`, `PO-6`, `GR-4`, `AG-3` |
| PO-4 | Rule for commands that print access keys | Done | MVP | `PR-2`, `PO-3` | `MV-1` |
| PO-5 | Inspect, explain, and validate policy | To do | Phase 2 | `PO-3`, `PO-6` | `AD-4` |
| PO-6 | Policy key for environment credentials | To do | Phase 2 | `PO-1`, `PO-3`, `PR-1` | `PO-5`, `FS-5` |

- Configuration validation shows up twice in the design. Validating on every load is part of `PO-1`. The standalone validate command is `PO-5`, which the design places in Phase 2.
- `PO-2` adds `aws-guard init` (default `config.toml` via `policy.DefaultFileBody`, optional `settings.toml` without loading it), and the wrapper cancels when the policy file is absent (`policy.Present` in `internal/gate`).
- `PO-3` adds `Evaluate` for profile policy: named profiles use `[profiles]` or `[defaults]`, and environment or unresolved subjects use `[defaults]` only in MVP. It returns the action and basis; grant and export checks are separate steps in `internal/gate`.
- `PO-4` adds `Invocation.PrintsAccessKeys` and `EvaluateExport` for `[credential_export]`. The wrapper applies export evaluation before profile policy and grants.
- `PO-6` implements the promise in [Environment-based credentials](DESIGN.md#64-environment-based-credentials) that `subject.Environment` (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`) has its own policy key in `config.toml` outside `[profiles]`, rather than only falling back to `[defaults]`. It introduces `config.toml` schema `version = 2` (while keeping `version = 1` supported in memory per [Versions and schema changes](DESIGN.md#versions-and-schema-changes)). When omitted, `subject.Environment` continues to use `[defaults]`. A grant still cannot authorize `subject.Environment`.

### GR: Grants

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| GR-1 | Grants file store | Done | MVP | `FS-1`, `FS-2` | `PR-5`, `GR-3`, `GR-4`, `GR-5`, `GR-7` |
| GR-2 | Credential-source fingerprint | Done | MVP | `PR-4` | `PR-5`, `GR-3`, `GR-4`, `GR-5` |
| GR-3 | `aws-guard allow` | Done | MVP | `GR-1`, `GR-2`, `FD-4` | `GR-6` |
| GR-4 | Checking grants (expiry and fingerprint) during policy evaluation | Done | MVP | `GR-1`, `GR-2`, `PO-3` | `GR-6`, `MV-1` |
| GR-5 | List grants | To do | Phase 3 | `GR-1`, `GR-2`, `PR-5` | — |
| GR-6 | Tests for concurrent allow and policy reload | To do | Phase 3 | `GR-3`, `GR-4` | — |
| GR-7 | `aws-guard revoke` | To do | Phase 2 | `GR-1`, `FD-4` | — |

- `GR-1` adds `internal/grants` (`Load`, `Update`, `Replace`, and corrupt-file diagnosis). A missing grants file means an empty list. The wrapper loads grants on every guarded invocation.
- `GR-2` adds `Fingerprint` and `Source.Equal` in `internal/grants` ([ADR 0012](decisions/0012-credential-source-fingerprint.md)). `GR-3` `aws-guard allow` calls `Fingerprint` and `Update`.
- `GR-3` adds `aws-guard allow` in `internal/cli`. It confirms the profile on a terminal and writes one grant.
- `GR-4` adds `grants.Current` and `policy.EvaluateGrant`. A `prompt` action runs only when a grant for that named profile is unexpired and the stored source matches the current fingerprint. `internal/gate` applies grant checking on every guarded invocation.
- `GR-7` is pulled ahead into Phase 2 and adds `aws-guard revoke <profile>` and `aws-guard revoke --all`, as specified in [Revoking grants](DESIGN.md#revoking-grants):
  - It needs no terminal, because it only narrows access.
  - It runs under the `grants.Update` lock.
  - With nothing to remove, it exits 0 with a note and writes nothing. A missing grants file is not created.
  - A corrupt grants file is refused with status 3, with the diagnosis note and a pointer to `aws-guard recover grants`. The file is not modified.
  - It pairs naturally with `GR-5`, so the user can see what to revoke, but neither blocks the other.

### MV: MVP milestone

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| MV-1 | Wrapper enforces policy end to end | Done | MVP | `PO-2`, `PO-4`, `GR-4`, `FS-3`, `PR-3` | — |
| MV-2 | Documentation of bypass limitations | Done | MVP | — | — |

### SS: SSO

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| SS-1 | How `aws sso login` is handled | To do | Phase 2 | `PR-2` | — |

- [ADR 0004](decisions/0004-login-out-of-scope.md) decides the behavior: recognize login and hand it to the real CLI, with no policy and no grant. `SS-1` is that recognition. Phase 2 remains the placeholder milestone.
- In `internal/gate/decide.go`, `gate.Decide` currently checks `policy.Present(env)`, resolves the subject, warns on `default`, and loads `config.toml` and `grants.json` before calling `invocation.Parse(args)`. Move `invocation.Parse(args)` and the login check to the top of `gate.Decide` so login works before `aws-guard init`, without a grant, and without printing a `default` profile warning.
- Pin down the exact argv rule for login in `internal/invocation` (for example `inv.OK && inv.Command == "sso" && inv.Operation == "login"` for the 2.9.0 floor; when `!inv.OK`, fail closed into the normal guarded path).

### CI: Continuous integration

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| CI-1 | macOS runner for tests and the native build | To do | Phase 4 | — | `AG-4` |

- Leave `CI-1` until `AG-4`. Ubuntu already runs the shared `unix` tests, and a `CGO_ENABLED=0` Darwin binary cross-compiles there. [ADR 0001](decisions/0001-linux-and-macos-first.md) covers platform-sensitive edits with a laptop `make check`. A standing `macos-15` job would add queue time on every pull request for that same source. Parent-process detection is the first planned code a green Ubuntu run will not compile: macOS has no `/proc`. Start `CI-1` as the first step of `AG-4`. If `OP-1` needs a Darwin-only session or terminal implementation sooner, start `CI-1` with that work instead.
- The macOS job runs `make test` and `make build` on the standard `macos-15` runner. Gitleaks, actionlint, ShellCheck, and govulncheck stay on `ubuntu-24.04`. That runner is free while this repository is public.

### AG: Agent awareness

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| AG-1 | Caller context, including whether a terminal is attached | To do | Phase 4 | — | `AG-2`, `AG-4`, `OP-1` |
| AG-2 | Environment-variable markers for agents | To do | Phase 4 | `AG-1` | `AG-3`, `AG-5` |
| AG-3 | Agent-specific policy rules | To do | Phase 4 | `AG-2`, `PO-3` | `FS-5`, `AD-3` |
| AG-4 | Parent-process detection | To do | Phase 4 | `AG-1`, `CI-1` | — |
| AG-5 | Tested integrations for selected agents | To do | Phase 4 | `AG-2` | `AD-5` |

- `AG-2` is the built-in environment markers from [ADR 0015](decisions/0015-initial-agent-integrations.md). They run only when `[agents.detection] environment_markers` is true ([ADR 0016](decisions/0016-agent-detection-switches.md)).
- `AG-3` applies `[agents]` actions only to a positive recognition. Those actions are present in the file, or they are not. The product does not add a default deny.
- `AG-4` inspects parent processes only when `[agents.detection] parent_process` is true.
- `AG-5` tests Claude Code, the OpenAI Codex CLI, Cursor, and Gemini CLI at pinned versions. Adding or dropping a product is a later change to that set.

### IN: Installation

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| IN-1 | Installer with manifest and `PATH` setup | To do | Phase 5 | — | `IN-2`, `IN-3`, `OP-2` |
| IN-2 | Shell integration | To do | Phase 5 | `IN-1` | — |
| IN-3 | Upgrade behavior | To do | Phase 5 | `IN-1`, `FS-5` | — |

- `IN-1` overlaps with the `install.sh` that the release pipeline in `OP-3` ships, as described in [Publish pipeline](DESIGN.md#123-publish-pipeline). They are related and may end up as one epic.
- `IN-1` records the Go wrapper path, the `aws` path, and the real AWS CLI path, and installs the POSIX disable script beside the installation ([ADR 0017](decisions/0017-persistent-disable-swaps-aws.md)). The enabled `aws` is the Go wrapper.
- `IN-2` is Bash and zsh, with shared POSIX `sh`, per [ADR 0005](decisions/0005-bash-and-zsh.md).

### OP: Operational criteria

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| OP-1 | Turn the wrapper off for one terminal session | To do | Backlog | `AG-1`, `FS-1`, `FD-4` | — |
| OP-2 | Persistent disable and enable | To do | Backlog | `IN-1` | — |
| OP-3 | GitHub Releases pipeline | To do | Backlog | — | `OP-4` |
| OP-4 | Auto-update and update notices | To do | Backlog | `FS-6`, `OP-3` | — |
| OP-5 | Tight and verbose logging | To do | Backlog | `FD-3` | — |

- These are the [recorded operational criteria](DESIGN.md#12-recorded-operational-criteria). The design keeps them out of Phases 1 to 6 until a later pass pulls them into a phase.
- `OP-2` is the swap in [ADR 0017](decisions/0017-persistent-disable-swaps-aws.md): disable, enable, the stand-in `aws`, and the marker file. `IN-1` has to lay down the script and the manifest fields first.

### AD: Advanced features

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| AD-1 | Identity verification through AWS STS | To do | Backlog | `FS-6`, `PR-5` | — |
| AD-2 | Audit logging | To do | Backlog | `FD-3` | — |
| AD-3 | Integration with OS-level isolation | To do | Backlog | `AG-3` | — |
| AD-4 | More detailed policy explanations | To do | Backlog | `PO-5` | — |
| AD-5 | Additional agent integrations | To do | Backlog | `AG-5` | — |

- These are Phase 6 in the design.
