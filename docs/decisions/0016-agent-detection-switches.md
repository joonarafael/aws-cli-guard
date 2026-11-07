# ADR 0016: Agent detection is three explicit switches

## Status

Accepted

## Context

Open question 9 in [DESIGN.md](../DESIGN.md) asked whether agent detection is on by default or opt-in. The design had bundled three different choices: consulting environment markers, applying agent allow and deny rules, and inspecting parent processes. The illustrative policy turned the first and the third on together and showed an `[agents]` default of deny.

Those choices have different effects. The built-in marker list is young and will change ([ADR 0015](0015-initial-agent-integrations.md)). Inherited variables can classify a human shell as an agent. Turning any of this on without an explicit choice would change who may skip session-off, and who is denied, before the user has asked for that.

Unknown callers already follow normal profile policy. A missing marker is not a human and is not an agent. Terminal state does not, by itself, select agent rules.

## Decision

Three responsibilities. Each is an explicit yes or no. One switch does not turn another on. A miss on one detector does not fall through to another. There is no silent default of on, and no invisible mode behind these keys.

1. **Environment markers are opt-in.** The feature ships with the built-in markers for the products in [ADR 0015](0015-initial-agent-integrations.md). Those markers are on only when the user sets `[agents.detection] environment_markers = true`. Until the user sets that, the wrapper does not consult agent markers, does not recognize a caller from them, and behaves as it does without this feature. `false` is the explicit off. When `[agents.detection]` is present, `environment_markers` is required and must be a boolean, so the table cannot exist without a yes or a no. When the table is absent, the user has not asked, and markers stay off. Init does not ask and does not write this key.

2. **Authorization is opt-in by presence.** Agent allow, prompt, and deny rules exist only when the policy file contains `[agents]` action keys (`default` or `[agents.profiles]`). Those rules apply only to a caller an enabled detector has positively recognized. The product does not invent a default deny for agents. Recognition with no action keys leaves allow, prompt, and deny on the normal profile policy. Action keys while both detectors are off make the file invalid: the file would claim a rule that cannot match. The load fails with a diagnosis that names the keys. It does not ignore the section.

3. **Parent-process detection is off unless set.** It runs only when `[agents.detection] parent_process = true`. Omission and `false` are off. `environment_markers = true` does not imply it. A missing marker does not select it. When the key is present it must be a boolean.

4. **What a match does.** A positive match from an enabled detector is a recognized agent. That status is what session-off refuses to exempt, what a denial may name as the caller, and what verbose logging may record. Those effects need the detector the user turned on. They do not need `[agents]` action keys. Allow, prompt, and deny change only when those action keys are present. With no detector on, no caller is recognized, so the session-off exception does not run.

Schema version 1 rejects `[agents]` ([ADR 0010](0010-policy-schema.md)). These keys arrive with the schema version that enforces them. User-declared extra markers stay explicit entries under `[agents.detection] user_declared` ([ADR 0023](0023-product-finish-before-agent-detection.md)). Declaring one does not turn the built-in set on. Turning the built-in set on does not invent markers the user did not write. Extra markers are consulted only along with the built-in set, and only when `environment_markers` is true. They are not a fourth switch.

## Consequences

- Existing policy files keep today's behavior when the feature ships.
- `AG-2` implements the built-in markers and the `environment_markers` switch. Shipping that code does not enable detection.
- `AG-3` implements `[agents]` actions for positive recognitions only.
- `AG-4` implements parent-process inspection only behind `parent_process = true`.
- Session-off cannot tell an agent from any other caller until the user turns a detector on. That is accepted while the marker list is young.
- A build that does not enforce `[agents]` still rejects a file that contains it.
