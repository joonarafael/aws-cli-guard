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

The MVP is complete. Two fixes to shipped behavior come first. Both are small, unblocked, and filed under Phase 2:

- `FD-6` skipping every copy of the wrapper during real AWS CLI discovery, so two copies cannot hand off to each other forever.
- `FD-5` resolving guard-owned paths from the user database, so an environment variable cannot point the wrapper at a forged policy or grants file ([ADR 0014](decisions/0014-guard-paths-from-user-database.md)).

Then the rest of Phase 2:

- `PR-5` detecting credential providers beyond the environment-key case.
- `SS-1` login handoff ([ADR 0004](decisions/0004-login-out-of-scope.md)).
- `PO-5` inspect, explain, and validate policy.

Also unblocked, but in later milestones: `GR-7`, `AG-1`, `IN-1`, `OP-3`.

`Q-6` and `Q-9` are answered, along with `Q-1`, `Q-2`, `Q-3`, and `Q-8`. Still open, and not in the way of the MVP: `Q-4`, `Q-5`, and `Q-7`.

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
| Phase 2: Better profile awareness | `FD-5`, `FD-6`, `PR-5`, `PR-6`, `PO-5`, `SS-1` |
| Phase 3: Grant operations | `FS-5`, `GR-5`, `GR-6`, `GR-7` |
| Phase 4: AI-agent awareness | `AG-1`, `AG-2`, `AG-3`, `CI-1`, `AG-4`, `AG-5` |
| Phase 5: Installation and portability | `IN-1`, `IN-2`, `IN-3` |
| Backlog | `FS-6`, `OP-1`, `OP-2`, `OP-3`, `OP-4`, `OP-5`, `AD-1`, `AD-2`, `AD-3`, `AD-4`, `AD-5` |

The MVP is complete when every MVP epic is `Done`.

## Open design decisions

Most of these come from [Open Design Questions](DESIGN.md#16-open-design-questions). The rest are decisions the design leaves open in other sections.

| ID | Decision | Status | Blocks |
| --- | --- | --- | --- |
| Q-1 | Which shells get first-class integration? | Answered in [ADR 0005](decisions/0005-bash-and-zsh.md) | — |
| Q-2 | Is `aws sso login` allowed by default, or does it get its own policy? | Answered in [ADR 0004](decisions/0004-login-out-of-scope.md) | — |
| Q-3 | Is identity verification optional, and when does it run? | Answered in [ADR 0006](decisions/0006-identity-verification-off-by-default.md) | — |
| Q-4 | Which AI agents are supported initially? | Open | `AG-5` |
| Q-5 | Is agent detection on by default, or opt-in? | Open | `AG-2` |
| Q-6 | Which minimum AWS CLI version and configuration features are supported? | Answered in [ADR 0002](decisions/0002-aws-cli-version-floor.md) | — |
| Q-7 | Which mechanism does persistent disable and enable use? | Open | `OP-2` |
| Q-8 | Is auto-update on or off by default? | Answered in [ADR 0007](decisions/0007-auto-update-asked-at-init.md) | — |
| Q-9 | Which exit status does the wrapper use for its own refusals? | Answered in [ADR 0003](decisions/0003-wrapper-exit-status.md) | — |

Q-1 to Q-6 are design questions 2, 4, 5, 8, 9, and 12. Q-7 comes from [Persistent system-wide disable and enable](DESIGN.md#122-persistent-system-wide-disable-and-enable), Q-8 from [Auto-update and update notices](DESIGN.md#124-auto-update-and-update-notices), and Q-9 from the exit-code requirement in [User Experience](DESIGN.md#9-user-experience). Q-4, Q-5, and Q-7 are the ones still open.

## Epics

### FD: Foundation

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| FD-1 | Wrapper entrypoint and real AWS CLI handoff | Done | MVP | — | `FD-6` |
| FD-2 | On-disk path resolution | Done | MVP | — | `FS-1`, `FD-5` |
| FD-3 | Wrapper exit status and secret-safe error messages | Done | MVP | — | `PR-3`, `FS-3`, `OP-5`, `AD-2` |
| FD-4 | `aws-guard` command skeleton and terminal confirmation | Done | MVP | — | `FS-4`, `PO-2`, `GR-3`, `GR-7`, `OP-1` |
| FD-5 | Guard-owned paths from the user database | To do | Phase 2 | `FD-2` | — |
| FD-6 | Skip every wrapper copy during real AWS CLI discovery | To do | Phase 2 | `FD-1` | — |

- `FD-5` implements [ADR 0014](decisions/0014-guard-paths-from-user-database.md) in `internal/paths`:
  - The home directory comes from `os/user.Current()`. `HOME` is the fallback when the lookup fails or returns an empty or non-absolute directory. The fallback is supported, prints no warning, and does not cancel.
  - `XDG_*` variables are no longer read.
  - The `AWS_CLI_GUARD_*_DIR` overrides compile only under the `awsguard_testpaths` build tag. `make test` and the integration-test build helper pass that tag; `make build` does not.
  - The lookup is injectable for unit tests.
  - Run `make check` on macOS before merging. Update [DEVELOPMENT.md](DEVELOPMENT.md) with the tag and the new dogfooding behavior.
- `FD-6` closes two self-handoff loops in `internal/executor/find.go`. Today `AWS_CLI_GUARD_REAL_AWS` is returned without a self-check, and `PATH` discovery only skips the running file. With two wrapper copies on `PATH`, each finds the other first. The design is in [Core Architecture](DESIGN.md#3-core-architecture):
  - A candidate is a wrapper when it is the same file as the running executable, or when `debug/buildinfo.ReadFile` reports the same main package path that `debug.ReadBuildInfo` reports for the running wrapper. A file with no Go build information is not a wrapper.
  - `PATH` discovery skips wrappers. An `AWS_CLI_GUARD_REAL_AWS` that names a wrapper cancels with status 3, without printing the path and without falling back to `PATH`.
  - Tests: a second wrapper copy earlier on `PATH` is skipped and the fake AWS runs once; the override naming the wrapper, or a copy of it, exits 3 and the fake does not run; the fake AWS, which is a different Go main package, is not treated as a wrapper.
  - Roughly 30 lines of code plus tests. It implements part of the Phase 5 safe-detection item early, the same way `PR-4` was pulled ahead.

### PR: Profile resolution

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| PR-1 | Credential subject classification | Done | MVP | `PR-2` | `PR-3`, `PO-3` |
| PR-2 | AWS CLI global option parsing | Done | MVP | — | `PR-1`, `PO-4`, `SS-1` |
| PR-3 | Warning when the `default` profile is used | Done | MVP | `PR-1`, `FD-3` | `MV-1` |
| PR-4 | Reading AWS config and credentials files, including custom locations | Done | MVP | — | `PR-5`, `GR-2` |
| PR-5 | Detecting credential providers: SSO, role chaining, `credential_process` | To do | Phase 2 | `PR-4` | `PR-6`, `AD-1` |
| PR-6 | Showing account and role in messages | To do | Phase 2 | `PR-5` | — |

- `PR-3` warns on stderr when the credential subject is the named profile `default`. The command still follows policy for `default`.
- `PR-4` is pulled ahead of Phase 2. The design lists custom AWS config locations under Phase 2, but the grant fingerprint in `GR-2` is MVP work and needs to read the AWS config file correctly.

### FS: File storage and recovery

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| FS-1 | Safe file handling: atomic writes, permissions, symlink refusal, locking | Done | MVP | `FD-2` | `FS-4`, `FS-5`, `FS-6`, `PO-1`, `GR-1`, `OP-1` |
| FS-2 | Strict, versioned file decoding | Done | MVP | — | `FS-3`, `FS-5`, `FS-6`, `PO-1`, `GR-1` |
| FS-3 | Diagnosis of corrupted files | Done | MVP | `FS-2`, `FD-3` | `FS-4`, `MV-1` |
| FS-4 | `aws-guard recover` | Done | MVP | `FS-1`, `FS-3`, `FD-4` | — |
| FS-5 | Migrate command for older file versions | To do | Phase 3 | `FS-1`, `FS-2` | `IN-3` |
| FS-6 | Preference settings file | To do | Backlog | `FS-1`, `FS-2` | `AD-1`, `OP-4` |

- `FS-6` is `settings.toml` from [ADR 0006](decisions/0006-identity-verification-off-by-default.md). Identity verification and the auto-update answer live there. `aws-guard recover settings` lands with that file. MVP `FS-4` covers policy, grants, and cache.
- `FS-3` adds `policy.Corruption` and `diagnose.Write` for policy loads. The wrapper prints the note after `output.Cancel` when a load is corrupt. `FS-4` adds `aws-guard recover policy`, `recover grants`, and `recover cache`. `aws-guard recover settings` stays with `FS-6`.

### PO: Policy

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| PO-1 | Policy schema and loading | Done | MVP | `FS-1`, `FS-2` | `PO-2`, `PO-3` |
| PO-2 | `aws-guard init` and the missing-policy flow | Done | MVP | `PO-1`, `FD-4` | `MV-1` |
| PO-3 | Policy evaluation engine | Done | MVP | `PO-1`, `PR-1` | `PO-4`, `PO-5`, `GR-4`, `SS-1`, `AG-3` |
| PO-4 | Rule for commands that print access keys | Done | MVP | `PR-2`, `PO-3` | `MV-1` |
| PO-5 | Inspect, explain, and validate policy | To do | Phase 2 | `PO-3` | `AD-4` |

- Configuration validation shows up twice in the design. Validating on every load is part of `PO-1`. The standalone validate command is `PO-5`, which the design places in Phase 2.
- `PO-2` adds `aws-guard init` (default `config.toml` via `policy.DefaultFileBody`, optional `settings.toml` without loading it), and the wrapper cancels when the policy file is absent (`policy.Present` in `internal/gate`).
- `PO-3` adds `Evaluate` for profile policy: named profiles use `[profiles]` or `[defaults]`, and environment or unresolved subjects use `[defaults]` only. It returns the action and basis; grant and export checks are separate steps in `internal/gate`.
- `PO-4` adds `Invocation.PrintsAccessKeys` and `EvaluateExport` for `[credential_export]`. The wrapper applies export evaluation before profile policy and grants.

### GR: Grants

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| GR-1 | Grants file store | Done | MVP | `FS-1`, `FS-2` | `GR-3`, `GR-4`, `GR-5`, `GR-7` |
| GR-2 | Credential-source fingerprint | Done | MVP | `PR-4` | `GR-3`, `GR-4`, `GR-5` |
| GR-3 | `aws-guard allow` | Done | MVP | `GR-1`, `GR-2`, `FD-4` | `GR-6` |
| GR-4 | Checking grants (expiry and fingerprint) during policy evaluation | Done | MVP | `GR-1`, `GR-2`, `PO-3` | `GR-6`, `MV-1` |
| GR-5 | List grants | To do | Phase 3 | `GR-1`, `GR-2` | — |
| GR-6 | Tests for concurrent allow and policy reload | To do | Phase 3 | `GR-3`, `GR-4` | — |
| GR-7 | `aws-guard revoke` | To do | Phase 3 | `GR-1`, `FD-4` | — |

- `GR-1` adds `internal/grants` (`Load`, `Update`, `Replace`, and corrupt-file diagnosis). A missing grants file means an empty list. The wrapper loads grants on every guarded invocation.
- `GR-2` adds `Fingerprint` and `Source.Equal` in `internal/grants` ([ADR 0012](decisions/0012-credential-source-fingerprint.md)). `GR-3` `aws-guard allow` calls `Fingerprint` and `Update`.
- `GR-3` adds `aws-guard allow` in `internal/cli`. It confirms the profile on a terminal and writes one grant.
- `GR-4` adds `grants.Current` and `policy.EvaluateGrant`. A `prompt` action runs only when a grant for that named profile is unexpired and the stored source matches the current fingerprint. `internal/gate` applies grant checking on every guarded invocation.
- `GR-7` adds `aws-guard revoke <profile>` and `aws-guard revoke --all`, as specified in [Revoking grants](DESIGN.md#revoking-grants):
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
| AG-2 | Environment-variable markers for agents | To do | Phase 4 | `Q-5`, `AG-1` | `AG-3`, `AG-5` |
| AG-3 | Agent-specific policy rules | To do | Phase 4 | `AG-2`, `PO-3` | `AD-3` |
| AG-4 | Parent-process detection | To do | Phase 4 | `AG-1`, `CI-1` | — |
| AG-5 | Tested integrations for selected agents | To do | Phase 4 | `Q-4`, `AG-2` | `AD-5` |

### IN: Installation

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| IN-1 | Installer with manifest and `PATH` setup | To do | Phase 5 | — | `IN-2`, `IN-3`, `OP-2` |
| IN-2 | Shell integration | To do | Phase 5 | `IN-1` | — |
| IN-3 | Upgrade behavior | To do | Phase 5 | `IN-1`, `FS-5` | — |

- `IN-1` overlaps with the `install.sh` that the release pipeline in `OP-3` ships, as described in [Publish pipeline](DESIGN.md#123-publish-pipeline). They are related and may end up as one epic.
- `IN-2` is Bash and zsh, with shared POSIX `sh`, per [ADR 0005](decisions/0005-bash-and-zsh.md).

### OP: Operational criteria

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| OP-1 | Turn the wrapper off for one terminal session | To do | Backlog | `AG-1`, `FS-1`, `FD-4` | — |
| OP-2 | Persistent disable and enable | To do | Backlog | `Q-7`, `IN-1` | — |
| OP-3 | GitHub Releases pipeline | To do | Backlog | — | `OP-4` |
| OP-4 | Auto-update and update notices | To do | Backlog | `FS-6`, `OP-3` | — |
| OP-5 | Tight and verbose logging | To do | Backlog | `FD-3` | — |

- These are the [recorded operational criteria](DESIGN.md#12-recorded-operational-criteria). The design keeps them out of Phases 1 to 6 until a later pass pulls them into a phase.

### AD: Advanced features

| ID | Title | Status | Milestone | Blocked by | Blocks |
| --- | --- | --- | --- | --- | --- |
| AD-1 | Identity verification through AWS STS | To do | Backlog | `FS-6`, `PR-5` | — |
| AD-2 | Audit logging | To do | Backlog | `FD-3` | — |
| AD-3 | Integration with OS-level isolation | To do | Backlog | `AG-3` | — |
| AD-4 | More detailed policy explanations | To do | Backlog | `PO-5` | — |
| AD-5 | Additional agent integrations | To do | Backlog | `AG-5` | — |

- These are Phase 6 in the design.
