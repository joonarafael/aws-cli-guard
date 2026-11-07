# ADR 0021: Policy file schema version 4 and agent action keys

## Status

Accepted

## Context

[ADR 0020](0020-policy-schema-v3.md) adds optional `[agents.detection]` with boolean `environment_markers` only. Positive recognition from markers does not change allow, prompt, or deny until agent action keys exist. [ADR 0016](0016-agent-detection-switches.md) requires those keys to be opt-in by presence and invalid when no detector is on.

Epic `AG-3` needs a schema version that accepts and enforces `[agents]` profile rules for recognized callers only.

## Decision

1. **Versions.** The binary accepts integer `version` 1, 2, 3, and 4. Version 5 and other unsupported values follow the strict decode path in `internal/decode`. The `aws` wrapper never migrates or rewrites an on-disk file. `aws-guard init` and `aws-guard recover policy` write version 2 when they create a fresh file.

2. **Versions 1 through 3.** Unchanged from [ADR 0010](0010-policy-schema.md), [ADR 0019](0019-policy-schema-v2.md), and [ADR 0020](0020-policy-schema-v3.md). Version 3 still rejects `[agents] default` and `[agents.profiles]` with `KindInvalid`. Keys `parent_process` and user-declared extra markers remain unknown until later epics.

3. **Version 4 document.** Version 3 plus optional agent authorization keys under `[agents]`:
   - Optional top-level `default` under `[agents]` with values `allow`, `prompt`, or `deny`.
   - Optional `[agents.profiles.<name>]` tables, each with required `action` (`allow`, `prompt`, or `deny`) and no other keys.
   - `[agents.detection]` unchanged from version 3.

4. **Validity.** Agent action keys are allowed only when `environment_markers = true`. That is the only detector this build implements. Action keys with the detection table absent, with `environment_markers = false`, or with no detection table fail validation and name `agents.default` when set, otherwise the lexicographically first `agents.profiles.<name>`. A file with detection only and no action keys stays valid; recognition then leaves profile policy unchanged.

5. **Evaluation.** For a positively recognized caller and a named profile subject, the wrapper uses `[agents.profiles.<name>]` when present, else `[agents] default` when present, else the normal profile policy. Environment and unresolved subjects ignore agent keys. There is no implied agent default of `deny`. Agent `deny` is evaluated before credential export and before grants. Agent `allow` and `prompt` still follow credential export, then grant checking for `prompt`.

6. **Production markers.** The built-in environment-marker catalog stays empty in production builds until pinned product variables ship (`AG-5`). Test builds may add a single test marker behind the `awsguard_testpaths` tag. The first shipped rows are [ADR 0022](0022-cursor-and-claude-code-markers.md).

7. **Inspection.** `aws-guard policy list` prints `agents_default` and each `agent_profile` / `agent_action` when those keys were present on load. `aws-guard policy explain` prints normal profile action and basis, and when a recognized caller would use an agent rule, `agent_action` and `agent_basis`.

## Consequences

- [ADR 0020](0020-policy-schema-v3.md) remains the version 3 record. Version 4 extends it here.
- `AG-4` adds `parent_process` when a schema change accepts it.
- The first built-in marker names are [ADR 0022](0022-cursor-and-claude-code-markers.md). Authorization uses this ADR once recognition is possible.
- `FS-5` `aws-guard migrate` may later rewrite older policy to version 4. The live `aws` path does not.
