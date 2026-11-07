# ADR 0024: `aws-guard migrate` command

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `FS-5` before installation and upgrades. The binary already loads older supported policy versions 1 through 3 and grants version 1, while enforcing version 4 policy and version 2 grants for new semantics. [DESIGN.md](../DESIGN.md) separates migration from recovery: migrate copies the live file aside and rewrites decisions in place; recover moves corrupt bytes aside and writes init defaults.

## Decision

1. **Command.** `aws-guard migrate` takes no arguments, requires no terminal confirmation, and is non-interactive so `IN-3` can run it from an upgrade path. Unknown arguments print `usage: aws-guard migrate` and exit 3.

2. **Scope.** The command migrates `config.toml` and `grants.json` when present. It does not read or write `settings.toml` or the cache.

3. **Preflight.** Both files are loaded before any aside copy. A missing policy file cancels with the existing missing-policy reason; nothing is written. An unusable policy or grants file cancels with corruption diagnosis; nothing is written. A missing grants file is normal; policy migration still runs.

4. **Targets.** Policy rewrites to integer version 4 (`readVersionAgentActions`). Grants rewrites to integer version 2 (`writeVersion` in `internal/grants`). Init and `aws-guard recover policy` still write policy version 2 via `DefaultFileBody`. The live `aws` path never calls migrate.

5. **Aside copy.** Before replacing a file that needs migration, the tool copies the current bytes to a sibling name `live.migrated.<utc-timestamp>` with the same numeric suffix rules as corrupt aside names. The live file stays in place until `files.WriteAtomic` replaces it. Recover continues to use `.corrupt.`.

6. **Idempotency.** A file already at the target version is left byte-for-byte unchanged and gets no aside copy. A second migrate is a no-op for that file.

7. **Policy rewrite.** Migrate encodes the loaded, validated `policy.Policy` into canonical version 4 TOML. It preserves actions, TTLs, environment table presence, and agent detection settings already on disk. It does not add `[agents]` or `environment_markers` when upgrading from version 1 or 2. It does not write `user_declared` or `parent_process`.

8. **Grants rewrite.** Migrate writes the loaded grant list through existing version 2 encoding. It does **not** call `grants.Fingerprint`. Stored sources remain the approvals already recorded. The user can run `aws-guard allow` again when live AWS config no longer matches.

9. **Locking.** Grants migrate holds the grants advisory lock across load, copy, and write. Policy migrate does not add a policy lock.

10. **Partial failure.** If policy migrate succeeds and grants migrate fails, exit 3 after reporting the policy aside path. The policy file is not rolled back.

11. **Stdout.** Detail lines report per-file outcomes: rewrite with from-version, aside path, already at target, or no grants file.

## Consequences

- [ADR 0018](0018-grants-schema-v2.md) consequence on fresh fingerprints applies only to explicit `allow`, not migrate.
- `IN-3` can invoke `aws-guard migrate` without a controlling terminal.
- When policy schema version 5 or later ships, migrate’s target version moves with the binary; this ADR’s mechanics stay the same.
