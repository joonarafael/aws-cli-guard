# ADR 0010: Policy file schema version 1

## Status

Accepted

## Context

MVP policy evaluation (`PO-3`) needs a strict, versioned user policy file at `config.toml`. [ADR 0008](0008-toml-decoder.md) provides strict TOML decoding. [DESIGN.md](../DESIGN.md) shows an illustrative policy block that mixes enforced sections with `[agents]`, which this build does not enforce yet. A file must not claim guarantees the binary does not provide.

The design leaves `approval_ttl` on profiles for `aws-guard allow` defaults. That field is not authorization by itself.

## Decision

1. **Version.** Only integer `version = 1` is accepted. Unsupported or missing versions follow the strict decode path in `internal/decode`.

2. **Actions.** Values are `allow`, `prompt`, and `deny`. Any other string, including empty, is invalid. Error messages name the key, not the value.

3. **Sections.**
   - `[defaults]`: omitted means default action `prompt`. Present table must include `action`.
   - `[credential_export]`: omitted means `prompt`. Present table must include `action`.
   - `[profiles]`: omitted means no profile overrides. Each profile table must include `action`.
   - `[agents]` and any other top-level table not in this schema are unknown and fail decode.

4. **Profile keys.** Names may contain spaces, dots, and other non-control characters (quote in TOML when needed). Empty names and names containing Unicode control characters are invalid. `[profiles.foo.bar]` is a nested table, not profile `foo.bar`; use `[profiles."foo.bar"]` for a dot in the name.

5. **Profile fields.** Only `action` and optional `approval_ttl` are allowed. `approval_ttl` is a positive Go duration string (`5m`, `1h30m`). Reject bare numbers, `0s`, negatives, and unit suffixes Go does not parse (for example `1d`). Valid on any profile action; the loader stores it for later allow commands and does not grant access.

6. **Rejected keys.** `classification`, `allow_readonly`, and other design examples that are not in this schema are unknown fields.

7. **Loading.** `internal/policy.Load` reads from an environment slice via `paths.ConfigFile`, `files.ReadRegular`, and `decode.Decode`. Maximum file size is 1 MiB. Missing file returns `ErrMissing` wrapping `os.ErrNotExist`. Guard-owned paths refuse a final symlink. Loose mode `0644` is not corruption. The wrapper does not call `Load` until evaluation exists.

## Consequences

- `PO-2` init writes a version-1 file matching this schema without `[agents]`.
- `PO-3` evaluates using exported `Policy`, `Profile`, and `Action` types.
- Agent policy requires a new schema version that defines `[agents]` and enforcement in the binary.
- The illustrative example in [DESIGN.md](../DESIGN.md) remains documentation; version 1 rejects `[agents]`.
