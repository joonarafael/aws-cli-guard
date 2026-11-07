# ADR 0023: Finish installation and operations before further agent detection

## Status

Accepted

## Context

Phases 1 through 3 are done. Caller context, the environment-marker switch, and `[agents]` action keys are done. The production catalog has Cursor and Claude Code ([ADR 0022](0022-cursor-and-claude-code-markers.md)). The OpenAI Codex CLI, Gemini CLI, and parent-process inspection are still open. Harnesses keep changing, so that catalog work does not finish.

The product still has no installer, no migrate command, no upgrade path, and none of the recorded operational criteria in [DESIGN.md](../DESIGN.md). Those criteria were left unscheduled on purpose. [ADR 0020](0020-policy-schema-v3.md) also left user-declared markers unnamed until a later decision.

Recognition is already one status. Policy, refusals, and the session-off exception key off a positive match. Another detector or another catalog row does not need a new engine.

## Decision

1. **Finish the product first.** Unfinished work follows this order. A later step waits even when its technical blockers are already done. Items in one step ship together.
   1. `FS-5`, `aws-guard migrate`.
   2. `IN-1`, the installer, and `OP-3`, the GitHub Releases pipeline.
   3. `IN-3`, upgrade behavior.
   4. `OP-1`, session off, then `OP-2`, persistent disable and enable.
   5. `OP-5`, tight and verbose logging, then `OP-4`, auto-update.
   6. `IN-2`, Bash and zsh integration.

2. **Agent detection last.** After that sequence, detection work continues in this order, and the catalog part continues as tools change:
   1. `AG-6`, user-declared markers.
   2. The rest of `AG-5`: the OpenAI Codex CLI and Gemini CLI at pinned versions. `AD-5` is the same kind of change for products after those four.
   3. `CI-1`, then `AG-4`, parent-process detection. The macOS runner starts with `AG-4` because that is the first code a green Ubuntu run will not compile.
   4. `AD-3`, OS-level isolation, after `AG-4`.

3. **Optional features stay optional.** `AD-1` (identity verification), `AD-4` (richer policy explanations), and `AD-2` (audit logging) follow a user need. They do not gate the sequence above.

4. **Detection stays opt-in while this work proceeds.** Init still writes policy version 2 and does not write `environment_markers`. A positive match is still the only recognized-agent status. New detectors set that same status. Session-off, verbose logging, and refusals key off it. A miss on one detector does not fall through to another.

5. **`user_declared`.** `AG-6` adds the key. The shape is fixed now so migrate and later schema versions have a name to aim at. Policy schema version 5 is version 4 plus this key. Versions 1 through 4 keep rejecting it.
   - The key is `user_declared`, under `[agents.detection]`.
   - The value is an array of environment variable names. Each name matches `^[A-Za-z_][A-Za-z0-9_]*$`. An empty string, a duplicate, or any other string fails validation and names `agents.detection.user_declared`.
   - The key is valid only when `environment_markers = true`. If the switch is false or `[agents.detection]` is absent, a present `user_declared` fails validation and names that key. Listing names does not turn the switch on. Turning the switch on does not invent names the user did not write.
   - When the switch is true, the wrapper consults the built-in catalog and these names together. A listed name matches when it is set to a non-empty value. The wrapper does not compare or log the value. Several matches are still one recognized caller.
   - A name that is also in the built-in catalog is allowed. The user list does not add a product to the set in [ADR 0015](0015-initial-agent-integrations.md).
   - An empty array is valid when the switch is true and adds no names. Absent `user_declared` means no extra names.
   - Init does not write the key. `aws-guard policy list` prints one `user_declared: <name>` line per name, in file order, when the key was present and non-empty.
   - Version 5 does not accept `parent_process` and does not change the rule that `[agents]` action keys require `environment_markers = true`. `AG-4` is a later schema version. That version accepts `parent_process` and widens the action-key rule to "at least one detector is on."

6. **Migrate.** `FS-5` rewrites an older supported policy file and an older supported grants file to the newest version that binary enforces. Today that is policy version 4 and grants version 2. It copies the current file aside first. It does not enable detection, does not write `environment_markers`, and does not write `user_declared`. The live `aws` path still never migrates. When version 5 or the parent-process schema ships, migrate learns that version and stays an explicit command.

## Consequences

- [ROADMAP.md](../ROADMAP.md) follows this order. Technical **Blocked by** links stay dependencies. This schedule is stricter.
- Codex CLI, Gemini CLI, parent-process inspection, and `user_declared` are not next. Until `AG-6`, a file that contains `user_declared` fails validation.
- `CI-1` still waits for `AG-4`.
- [ADR 0016](0016-agent-detection-switches.md) still separates the switches. `user_declared` is part of the environment-marker switch, not a fourth switch.
