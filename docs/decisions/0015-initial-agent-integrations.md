# ADR 0015: Initial agent integrations are four products

## Status

Accepted

## Context

Open question 8 in [DESIGN.md](../DESIGN.md) asked which AI agents the wrapper supports initially. Detection is heuristic. There is no universal marker, and a name that is wrong or ambient will misclassify a human terminal once detection is turned on.

The design listed candidate variables, not a committed set. Coding agents, the harnesses around them, IDEs, and the models those products wrap change constantly. A list written down today will not stay accurate.

## Decision

1. **Initial set.** The built-in integrations are Claude Code, the OpenAI Codex CLI, Cursor, and Gemini CLI. Other candidates, including OpenCode, Goose, and the Copilot CLI, are outside this set.

2. **The set changes.** This is a starting point, not a permanent contract. Adding a product, dropping one, or changing how a listed product is recognized is expected. That happens as a later decision when the integration is actually checked. It does not wait for the list to become complete.

3. **Products, not frozen variable names.** This decision does not name environment variables. Each integration records the command-scoped variables set by one pinned version of that product, and it records the version. A variable the pinned version does not set is not shipped under that product's name. A variable that product also injects into an ordinary human terminal is not the signal for it.

4. **User-declared markers stay extra.** Markers the user writes are explicit additions. They do not enlarge this built-in set, and this set does not imply markers the user did not write.

## Consequences

- `AG-5` tests these four products at pinned versions. A later product is further work (`AD-5` or a change to this set), not part of the first integration pass. The first recorded variables are [ADR 0022](0022-cursor-and-claude-code-markers.md). The remaining two products wait until the detection tail in [ADR 0023](0023-product-finish-before-agent-detection.md).
- User-declared additions use the policy key `user_declared` ([ADR 0023](0023-product-finish-before-agent-detection.md)). That key does not enlarge this set.
- An inherited marker can outlive the agent and classify a later shell as that agent. The integration documents the inheritance it observed on the pinned version.
- The candidate variable list in earlier drafts of the design is retired. The current text points here.
