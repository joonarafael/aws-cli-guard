# ADR 0034: Verified AWS CLI ceiling 2.40.3

## Status

Accepted

## Context

The wrapper parses top-level global options from argv before policy runs. Unknown options that take a value can shift positional tokens so command and operation are misread, which weakens rules such as `[credential_export]`. The verified ceiling in [ADR 0002](0002-aws-cli-version-floor.md) must move only when a maintainer reviews upstream changes and updates the parser.

AWS CLI 2.34.0 added global `--cli-error-format` and `off` as an `--output` choice. 2.39.6 added a global `--help` flag. 2.32.0 added `aws login` and the `login_session` profile key. Users on 2.36.x are common; the latest v2 release at the time of this change is 2.40.3.

## Decision

1. **Ceiling.** Raise the verified ceiling to **2.40.3**. Constants live in `internal/cliversion/version.go`. Warning text uses the floor and ceiling constants, not hard-coded version strings.

2. **Parser.** Extend `internal/invocation` global options through 2.40.3:
   - `--cli-error-format` with choices `legacy`, `json`, `yaml`, `text`, `table`, `enhanced`.
   - `off` added to `--output` choices.
   - `--help` as a flag option.

3. **Verification record.** "Verified" here means changelog review from 2.9.0 through 2.40.3, parity with `awscli/data/cli.json` for global options, and tests for the parser additions. A maintainer ran the wrapper against local fake CLIs reporting 2.36.14 and 2.40.3 during the release hardening pass.

4. **Known gaps (unchanged behavior).**
   - `aws login` is guarded like any other command ([ADR 0004](0004-login-out-of-scope.md) covers `aws sso login` only).
   - Grant fingerprints do not include `login_session` ([ROADMAP.md](../ROADMAP.md) `PR-7`).
   - `aws update` is guarded like any other command.

5. **Follow-ups.** [ROADMAP.md](../ROADMAP.md) adds `SS-2` (`aws login` handling) and `PR-7` (`login_session` in the credential-source fingerprint).

## Consequences

- [DEVELOPMENT.md](../DEVELOPMENT.md) next free ADR number is `0035`.
- A CLI newer than 2.40.3 warns once per session and still runs policy until the ceiling moves again.
- Future ceiling bumps must include the global-option table and a new ADR entry.
