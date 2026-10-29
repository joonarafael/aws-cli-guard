# ADR 0013: Credential-export command classification

## Status

Accepted

## Context

MVP policy evaluation (`PO-4`) must enforce `[credential_export]` from [ADR 0010](0010-policy-schema.md). [DESIGN.md](../DESIGN.md) names a fixed set of AWS CLI commands that print access keys. Profile `allow` and system-wide grants must not override that rule. The wrapper must not retain setting names or argv fragments that could echo secrets in later messages.

`internal/invocation` already parses top-level global options and records command and operation. It does not yet classify credential-export invocations.

## Decision

1. **Classification API.** After a successful parse, `Invocation` carries `PrintsAccessKeys bool`. The setting name for `configure get` is not stored on the struct. Classification uses the full positional slice after command and operation, before those fields are copied into `Invocation`.

2. **Failed parse.** When global-option parsing fails, `PrintsAccessKeys` is false. Credential-export classification does not run.

3. **In the set.**
   - `configure` + `export-credentials`, including further positional tokens (for example subcommand flags).
   - `configure` + `get` with zero remaining positionals, or two or more: treat as secret-printing (fail-closed).
   - `configure` + `get` with exactly one remaining positional: split on the first `.`. The segment after the first dot, or the whole token when there is no dot, is the setting name. The command is in the set only when that name is exactly `aws_access_key_id`, `aws_secret_access_key`, or `aws_session_token`. Comparison is case-sensitive. Do not read AWS config files.

4. **Outside the set.** `configure list-profiles`, `configure set`, `sso login`, ordinary service commands, and `configure get` with a known non-secret setting (for example `region`).

5. **Policy gate.** `internal/policy.EvaluateExport` applies `Policy.CredentialExport` when `PrintsAccessKeys` is true. `allow` permits the gate to pass; profile policy still runs later. `prompt` and `deny` refuse. `prompt` here does not consult grants. Unknown stored actions are treated as `deny`. When `PrintsAccessKeys` is false, the export decision does not apply.

6. **Refusal text.** One fixed cancellation reason, with no command or setting text: `commands that print access keys are refused until credential_export action is allow`.

7. **Wrapper.** `cmd/aws` does not call `EvaluateExport` until `MV-1`. Grant checking stays in `GR-4`.

## Consequences

- `PO-4` adds `internal/invocation/export.go` and `internal/policy/export.go`.
- `MV-1` runs export evaluation before profile policy and grants.
- Extending the command set requires an intentional code and documentation change, not config alone.
