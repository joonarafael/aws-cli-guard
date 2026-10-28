# ADR 0011: Grants file schema version 1

## Status

Accepted

## Context

MVP grants (`GR-3`, `GR-4`) need a strict, versioned grants file at `grants.json` under the state directory. [ADR 0008](0008-toml-decoder.md) provides strict JSON decoding. [DESIGN.md](../DESIGN.md) says a grant record stores the profile name, the expiry, and a fingerprint of the credential source: account, role ARN, SSO account and role, and whether `credential_process` is set. The fingerprint must not include secrets or the credential-process command.

[ADR 0009](0009-aws-config-read.md) leaves filling that fingerprint to `GR-2`. A missing grants file means there are no grants. An expired grant is not corruption. `source_profile`, `credential_source`, and role chains are Phase 2 (`PR-5`).

## Decision

1. **Version.** Only integer `version` 1 is accepted. Unsupported or missing versions follow the strict decode path in `internal/decode`.

2. **Document.** A present file is a JSON object with `version` and `grants`. `grants` is required. An empty array grants nothing. Omitting `grants`, or JSON `null`, is invalid. A missing file is not a document: `Load` returns version 1 and an empty grant list, creates nothing, and returns no error.

3. **Grant fields.** Each element has `profile`, `expires`, and `source`. All three are required. Unknown fields are errors.

4. **Profile.** Same rules as policy profile names in [ADR 0010](0010-policy-schema.md): non-empty, Unicode control characters invalid, spaces and dots allowed. At most one grant per profile. A repeated profile name is invalid. The key in the error identifies the second element; the message does not repeat the name.

5. **Expiry.** `expires` is an RFC3339 timestamp. Offsets are accepted and stored in UTC. Resolution is one second; a fractional second is truncated. The value is not checked against the clock. An unparseable timestamp is invalid. The error names the key and does not include the raw text.

6. **Source.** `source` is a required object. Its only keys are `account_id`, `role_arn`, `sso_account_id`, `sso_role_name`, and `credential_process`. The string keys are required and may be empty. `credential_process` is a required boolean, not the command string. `account_id` is the design's account field. This package does not derive one field from another. A grant does not record `source_profile`, `credential_source`, or a role chain. Those stay out until a later schema version.

7. **Loading and writing.** `internal/grants.Load` reads from an environment slice via `paths.GrantsFile`, `files.ReadRegular`, and `decode.Decode`. Maximum file size is 1 MiB. Guard-owned paths refuse a final symlink. Loose mode `0644` is not corruption. `Update` holds `files.OpenLock` across load, the caller mutation, and `files.WriteAtomic`. `Replace` holds the lock and writes without reading. Writes create the state directory at `0700` and the file at `0600`. The encoded file is version 1, two-space indented, with a trailing newline, every source key present, and `expires` in UTC `Z` form. The wrapper does not call this package until evaluation exists.

## Consequences

- `GR-2` fills `source` from `internal/awsconfig` and defines when two sources match. It does not change this schema.
- `GR-3` `aws-guard allow` mutates grants through `Update`.
- `GR-4` decides expiry and source match. `Load` returns expired grants.
- `FS-4` `aws-guard recover grants` moves the unusable file aside and `Replace`s it with version 1 and an empty list.
- Adding a bound on `source_profile` or role chaining is a new schema version.
