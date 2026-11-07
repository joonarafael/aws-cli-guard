# ADR 0020: Policy file schema version 3 and agent detection switch

## Status

Accepted

## Context

[ADR 0019](0019-policy-schema-v2.md) defines policy file versions 1 and 2. Phase 4 (`AG-2`) adds built-in environment-marker detection behind an explicit user switch ([ADR 0016](0016-agent-detection-switches.md)). The initial product set is [ADR 0015](0015-initial-agent-integrations.md). Variable names are recorded only with a pinned product version; this ADR does not name them.

Versions 1 and 2 reject any `[agents]` table so a file cannot claim agent guarantees the binary does not enforce. Detection settings need a schema version that accepts only what this build implements.

## Decision

1. **Versions.** The binary accepts integer `version` 1, 2, and 3. Version 4 and other unsupported values follow the strict decode path in `internal/decode`. The `aws` wrapper never migrates or rewrites an on-disk file. `aws-guard init` and `aws-guard recover policy` write version 2 when they create a fresh file.

2. **Version 1 and 2 documents.** Unchanged from [ADR 0010](0010-policy-schema.md) and [ADR 0019](0019-policy-schema-v2.md). If the file contains a top-level `[agents]` table, validation fails with kind `KindInvalid` and key `agents`. The error does not echo table values.

3. **Version 3 document.** Version 2 plus an optional top-level `[agents]` table. Only `[agents.detection]` is allowed under `[agents]`. When `[agents.detection]` is present, `environment_markers` is required and must be a boolean. When `[agents.detection]` is absent, including an empty `[agents]` table, markers stay off (`HasEnvironmentMarkers` false). Keys `default`, `[agents.profiles]`, and `parent_process` are unknown until later epics enforce them; they fail decode on version 3.

4. **Matching.** When the loaded policy has `environment_markers = true`, the wrapper consults a built-in catalog of environment variable names. A name matches when it appears in the invocation environment slice with a non-empty value (first-wins, case-sensitive). The wrapper does not compare or log values. Several matches yield one recognized agent caller. An empty built-in catalog recognizes nobody until rows are added with a pinned product version. User-declared extra markers are not part of version 3. The key `user_declared` is specified in [ADR 0023](0023-product-finish-before-agent-detection.md) and stays unknown until policy schema version 5.

5. **Authorization unchanged in this version.** Positive recognition does not change allow, prompt, or deny. Agent action keys arrive with a later schema version (`AG-3`).

6. **Inspection.** `aws-guard policy list` prints `environment_markers: true` or `environment_markers: false` after `credential_export` when `[agents.detection]` was present on load. It omits that line when the detection table was absent.

## Consequences

- [ADR 0019](0019-policy-schema-v2.md) remains the version 2 record. Version 3 extends it here.
- `AG-3` adds `[agents]` action keys and evaluation for recognized callers only.
- `AG-4` adds `parent_process` behind the same detection table when a schema change accepts it.
- The first built-in catalog rows are [ADR 0022](0022-cursor-and-claude-code-markers.md). Those rows do not change profile authorization until `AG-3`.
- `FS-5` `aws-guard migrate` rewrites an older supported policy file to the newest version that binary enforces. It does not enable detection and does not write `user_declared` ([ADR 0023](0023-product-finish-before-agent-detection.md)). The live `aws` path does not migrate.
