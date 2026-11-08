# ADR 0027: Session off for one terminal session

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `OP-1` after upgrade and before persistent disable. [DESIGN.md](../DESIGN.md) specifies session off in [Turn the wrapper off for one terminal session](../DESIGN.md#121-turn-the-wrapper-off-for-one-terminal-session). Caller context and recognized-agent status already exist ([ADR 0016](0016-agent-detection-switches.md)). Session identity for caches uses `getsid` in `internal/session` ([ADR 0002](0002-aws-cli-version-floor.md)). Shell integration is not required ([ADR 0005](0005-bash-and-zsh.md)). A child `aws-guard off` cannot export state back to the parent shell ([ADR 0002](0002-aws-cli-version-floor.md)), so honor cannot depend on an environment variable.

## Decision

1. **Commands.** `aws-guard off` and `aws-guard on` take no arguments. Extra arguments exit 3 with a usage line and write nothing.

2. **Confirmation.** `aws-guard off` requires the controlling terminal and confirmation phrase `off`. A headless caller or a mismatch exits 3 with `the session change was not confirmed` and does not touch the file. `aws-guard on` does not confirm and does not need a terminal, same posture as `aws-guard revoke`.

3. **File.** `sessions.json` lives beside `grants.json` under the state directory ([On-disk files](../DESIGN.md#on-disk-files)). Writes use `files.OpenLock`, append one record, and `files.WriteAtomic`. Mode `0600`, directory `0700`. A missing file means no records. The live `aws` path never creates the file. Corrupt data cancels with status 3 and points at `aws-guard recover sessions`. It is never treated as off.

4. **Schema version 1.** Strict JSON through `internal/decode`. Required top-level `version`. Required `records` array. Unknown fields and duplicate keys fail. Latest record for a session is the last array element with that `session_id`, not the greatest timestamp. Records are not pruned.

5. **Record fields.** Each record has `session_id` (`sid:<positive>` or `ppid:<digits>`, matching `session.ID()`), `effect` (`off` or `on`), `token`, `at` (RFC3339 UTC), and optional `tty`. `tty` holds `path`, `dev`, `ino`, and `rdev` as JSON integers. An `off` record must include `tty`. An `on` record may omit `tty`.

6. **Token (option A).** `off` and `on` each generate a new token with `crypto/rand` (32 bytes, 64 lowercase hex digits). The reader requires that shape. It does not compare the token to the environment, a keyring, or any second file. The token is not a well-known constant. A same-user writer of the state directory can append a well-formed record; that is the same class of bypass as editing `config.toml` ([Bypass Limitations](../DESIGN.md#10-bypass-limitations)). No environment variable selects session off.

7. **Honor rule.** After policy, settings, and grants load and after caller observation, when the caller is not recognized, load `sessions.json`. Honor the latest record for `session.ID()` only when all hold:
   - `effect` is `off` and `token` is well-formed
   - the record includes `tty`
   - if this process has a controlling terminal, its path and device numbers match the record
   - if not, `Lstat` of the recorded path is a character device that is not a symlink and the device numbers match
   A new terminal that reuses a session id has a different device, so an old record does not apply. A child with no terminal stays off only while that terminal node still exists. A later `on` for the same session id restores policy. Recognized agents ignore session off ([ADR 0016](0016-agent-detection-switches.md)). `aws sso login` still returns before any session read ([ADR 0004](0004-login-out-of-scope.md)).

8. **Placement.** The skip runs in `gate.decideAfterLoads` after `observeCaller` and before export and profile evaluation. Skipped invocations print one `aws-cli-guard: warning:` line: `this terminal session is off and policy is not applied`. Refusals do not mention the mechanism, path, token, or real AWS CLI path.

9. **Recover.** `aws-guard recover sessions` confirms with phrase `sessions`, moves the file aside, and writes version 1 with an empty `records` array. A missing file is not created.

10. **Out of scope.** `init`, `migrate`, and `upgrade` do not read or write this file. Persistent disable (`OP-2`) is separate ([ADR 0017](0017-persistent-disable-swaps-aws.md)).

## Consequences

- [ROADMAP.md](../ROADMAP.md) marks `OP-1` done; `OP-2` is next in the product-finish sequence.
- The on-disk layout table in [DESIGN.md](../DESIGN.md#on-disk-files) includes `sessions.json`.
- macOS tty capture uses `fcntl(F_GETPATH)`; platform-sensitive edits should pass `make check` on a Mac ([DEVELOPMENT.md](../DEVELOPMENT.md)).
- [DEVELOPMENT.md](../DEVELOPMENT.md) next ADR number is `0028`.
