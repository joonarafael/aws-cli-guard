# ADR 0012: Credential-source fingerprint

## Status

Accepted

## Context

[ADR 0011](0011-grants-schema.md) defines the five fields stored in each grant's `source` object. [ADR 0009](0009-aws-config-read.md) reads and merges AWS shared config and credentials files without returning secrets. `GR-2` must fill `grants.Source` from a named profile and define when two sources match, without changing the grants file schema or following role chains.

## Decision

1. **API.** `internal/grants.Fingerprint(env, profile)` loads AWS config via `awsconfig.Load`, reads `Profiles[profile]` only, and returns a `Source`. `Source.Equal` compares all five fields. There is no digest or hash.

2. **Empty profile.** An empty profile name returns `ErrEmptyProfile` and does not read files.

3. **Missing profile.** When the profile is absent from the merged map, or both AWS files are missing, return an empty `Source` and no error. That is a readable config with no fingerprint fields.

4. **Unreadable config.** Any `awsconfig.Load` error is returned unchanged. Callers must not treat a read or parse failure as an empty fingerprint.

5. **Field mapping.** Copy top-level settings only: `account_id`, `role_arn`, `sso_account_id`, and `sso_role_name` from the merged section. Set `credential_process` from `Section.HasCredentialProcess`. Do not parse `role_arn` for an account ID. Do not copy `source_profile`, `credential_source`, or keys from `[sso-session ...]` blocks. Do not walk chains.

6. **Validation.** Copied strings must pass the same no-control-character rule as loaded grants. Failure returns `*grants.Error` with `KindInvalid` and a `source.*` key. Values are not echoed.

7. **Wrapper.** `cmd/aws` and `cmd/aws-guard` do not call `Fingerprint` until `GR-3` and `GR-4`.

## Consequences

- `GR-3` `aws-guard allow` calls `Fingerprint` before writing a grant.
- `GR-4` compares stored and current sources with `Equal` when checking grants.
- Role chains and `sso-session` resolution belong in `PR-5` and would require a new grants schema version to record.
