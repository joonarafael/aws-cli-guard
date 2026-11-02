# ADR 0019: Policy file schema version 2 and environment credentials

## Status

Accepted

## Context

[ADR 0010](0010-policy-schema.md) defines policy file version 1. Phase 1 treats `subject.Environment` (static credentials from `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`) like `[defaults]` for profile policy. [DESIGN.md](../DESIGN.md) says environment credentials are a separate subject that can be named in policy outside `[profiles]`. A named-profile grant must not authorize them.

Phase 2 (`PO-6`) adds that key without changing grant semantics or adding migration on the live `aws` path.

## Decision

1. **Versions.** The binary accepts integer `version` 1 and 2. Version 3 and other unsupported values follow the strict decode path in `internal/decode`. The `aws` wrapper never migrates or rewrites an on-disk file. `aws-guard init` and `aws-guard recover policy` write version 2 when they create a fresh file.

2. **Version 1 document.** Unchanged from [ADR 0010](0010-policy-schema.md) for `[defaults]`, `[credential_export]`, and `[profiles]`. If the file contains a top-level `[environment]` table, validation fails with kind `KindInvalid` and key `environment`. The error does not echo the action value.

3. **Version 2 `[environment]`.** Optional top-level table. Omitted means `subject.Environment` uses `[defaults]` for profile policy, same as version 1. Present table must include `action` with values `allow`, `prompt`, or `deny`. Only `action` is allowed in the table. Unknown keys inside the table or other unknown top-level tables fail decode. `[agents]` is still unknown until a later schema version enforces it.

4. **Evaluation.** `policy.Evaluate` applies `[environment]` only to `subject.Environment` when the table was present on load (`HasEnvironmentAction`). `subject.Unresolved` always uses `[defaults]`. Named profiles are unchanged. Grants still apply only to named profiles; a grant cannot authorize `subject.Environment`.

5. **Basis.** When `[environment]` is present and applies, the decision basis is `BasisEnvironment`. When omitted, environment credentials use `BasisDefault`.

6. **Init default.** `DefaultFileBody` is version 2 and includes `[environment] action = "prompt"` alongside `[defaults]` and `[credential_export]`, so the key is visible in new files.

## Consequences

- [ADR 0010](0010-policy-schema.md) remains the version 1 record. Version 2 extends it here.
- `PO-5` `policy explain` and `config validate` can describe the environment rule once they land.
- `FS-5` `aws-guard migrate` may later rewrite version 1 policy to version 2. The live `aws` path does not.
