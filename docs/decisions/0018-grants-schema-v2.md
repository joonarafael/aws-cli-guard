# ADR 0018: Grants file schema version 2 and credential-chain fingerprint

## Status

Accepted

## Context

[ADR 0011](0011-grants-schema.md) defines grants file version 1. [ADR 0012](0012-credential-source-fingerprint.md) fills `source` from the named profile’s top-level settings only. A grant for a profile with `source_profile`, `credential_source`, or `sso_session` does not bind changes to parent profiles, ambient sources, or `[sso-session]` blocks. Phase 2 (`PR-5`) needs that binding without running `credential_process`, calling AWS, or storing secrets.

[ADR 0009](0009-aws-config-read.md) already merges AWS shared config and credentials files. Chain walking and SSO-session lookup extend read-only resolution in `internal/awsconfig`.

## Decision

1. **Versions.** The binary accepts integer `version` 1 and 2. Version 3 and other unsupported values follow the strict decode path. The `aws` wrapper never migrates or rewrites an on-disk file. `aws-guard allow`, `Replace`, and `recover grants` write version 2. A missing grants file is version 2 with an empty grant list.

2. **Version 1 source.** Unchanged from [ADR 0011](0011-grants-schema.md): `account_id`, `role_arn`, `sso_account_id`, `sso_role_name`, and `credential_process` only. Unknown keys are errors.

3. **Version 2 source.** The outer `source` object keeps the version 1 keys and adds required strings `source_profile`, `credential_source`, `sso_session`, `sso_start_url`, `sso_region`, and `external_id`. Each may be empty. `credential_process` remains a boolean, not a command. Required `chain` is a JSON array; it may be empty. Each element is an object with required `profile` and the same credential string keys and `credential_process` as the outer object, with no nested `chain`. Unknown keys are errors. JSON `null` for any required key is invalid.

4. **Resolution.** `internal/awsconfig.ResolveProfile` walks from a named profile in the merged config:
   - Profile lookup is an exact key match in `Profiles`.
   - When both `source_profile` and `credential_source` are non-empty on the same section, resolution fails.
   - When `credential_source` is non-empty, it is stored and no upstream walk runs.
   - When `source_profile` is non-empty, upstream hops are recorded nearest parent first. A non-empty `source_profile` is followed even when `role_arn` is empty. The starting profile is in the visited set before the walk; re-entering a profile is a cycle. More than eight upstream hops fails.
   - When `sso_session` is non-empty on a section, `sso_start_url` and `sso_region` come from that `[sso-session]` block only. When `sso_session` is empty, those keys come from the section. `sso_account_id` and `sso_role_name` always come from the section. A missing `[sso-session]` block fails resolution.
   - `external_id` is copied from the section. Web identity, `mfa_serial`, `role_session_name`, `duration_seconds`, and generic `region` are not stored. `account_id` is not derived from `role_arn`.

5. **Fingerprint.** `grants.Fingerprint` calls `awsconfig.Load` and `ResolveProfile`. A missing named profile returns an empty `Source` and no error. Load and parse errors propagate. Resolution failures return `*grants.Error` with kind `KindResolve`. String values pass the same no-control-character rule as version 1; failures use `KindInvalid` and do not echo values.

6. **Compare.** `Source.Equal` compares every outer field and every hop in order. A nil or empty `chain` matches an empty chain. `grants.Current` does not branch on file version. A version 1 grant loads with empty version 2 fields, so plain role, inline SSO, and `credential_process` grants still match. A version 1 grant does not match once the live profile has a non-empty `source_profile`, `credential_source`, or `sso_session` until the user runs `aws-guard allow` again. `Update` rewrites the whole file as version 2 and copies other grants as stored without re-fingerprinting them.

7. **Errors.** `awsconfig` resolution failures use kind `KindResolve`. The error key is one of `cycle`, `depth`, `source_profile`, `credential_source`, or `sso_session`. Messages do not include profile names, URLs, or other config values. `aws-guard allow` maps `KindResolve` to `the credential source could not be resolved`. Other fingerprint failures keep `the aws config could not be read`.

8. **Encode.** Version 2 files are two-space indented JSON with a trailing newline. Every required source and hop key is present on write.

## Consequences

- [ADR 0011](0011-grants-schema.md) remains the version 1 record. Version 2 extends it here.
- [ADR 0012](0012-credential-source-fingerprint.md) remains the version 1 fill rule. Chain and SSO-session binding live in this ADR and `ResolveProfile`.
- `PR-6` refusal detail lines and `PO-5` `policy explain` can reuse `ResolveProfile`.
- `FS-5` `aws-guard migrate` rewrites version 1 grants to version 2 with stored sources unchanged ([ADR 0024](0024-migrate-command.md)). The live `aws` path does not.
