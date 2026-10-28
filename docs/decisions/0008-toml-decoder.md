# ADR 0008: TOML decoder for strict file decoding

## Status

Accepted

## Context

[Versions and schema changes](../DESIGN.md#versions-and-schema-changes) require strict TOML decoding for user policy and `settings.toml`, and strict JSON for grants. Unknown fields, duplicate keys, and a required integer `version` must be errors. Corrupt-file diagnosis must not echo file bodies or parser snippets that might contain pasted credentials.

The repository keeps product dependencies minimal. [`.golangci.yml`](../../.golangci.yml) `depguard` allows only the standard library and this module.

Go has no standard TOML package. JSON uses `encoding/json`, but that decoder keeps the last value when a key is repeated, so duplicate-key rejection for grants needs a separate scan in this module.

## Decision

1. **TOML library.** Product code may import `github.com/pelletier/go-toml/v2` at the version pinned in [go.mod](../../go.mod). Imports must use the public API only, not packages under `unstable`.

2. **JSON.** Product code uses `encoding/json` for the final unmarshal into caller structs, plus token scanning in `internal/decode` to detect duplicate keys and to classify the top-level `version` before strict unmarshal.

3. **Error safety from go-toml.** When mapping parser failures into `internal/decode` errors, copy only `DecodeError.Key()` and `DecodeError.Position()`. Do not call `DecodeError.String()` or copy library `Error()` text into user-facing messages; type and syntax messages can quote literals from the document.

4. **Further dependencies.** Any additional third-party import in product code needs its own ADR and a matching `depguard` allow entry.

## Consequences

- `FS-2` and later `PO-1` / `GR-1` share one strict decode path for TOML and JSON.
- Security review of TOML parsing tracks the pinned `go-toml/v2` release and our wrapper, not ad hoc string handling of file contents in errors.
- Bumping `go-toml/v2` is a normal dependency change; behavior changes should be covered by `internal/decode` tests.
