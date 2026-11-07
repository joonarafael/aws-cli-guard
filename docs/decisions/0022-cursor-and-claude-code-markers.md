# ADR 0022: First environment markers are Cursor and Claude Code

## Status

Accepted

## Context

[ADR 0015](0015-initial-agent-integrations.md) names the initial products and leaves each variable to the integration that checks it. [ADR 0020](0020-policy-schema-v3.md) matches a built-in catalog only when `environment_markers = true`. [ADR 0021](0021-policy-schema-v4.md) keeps that catalog empty until pinned product variables ship. A non-empty value is enough to match. The wrapper does not compare or log the value.

The OpenAI Codex CLI and Gemini CLI have not been checked. `CLAUDECODE` is the other Claude Code name on the vendor variables page. That page says IDE extensions set it in their integrated terminals.

## Decision

1. **Catalog rows.** Production builds recognize one agent caller when either name is set to a non-empty value and `environment_markers = true`. The names are `CURSOR_AGENT` and `CLAUDE_CODE_CHILD_SESSION`. Several matches stay one recognized caller. Test builds still append `AWS_CLI_GUARD_TEST_AGENT` behind the `awsguard_testpaths` tag.

2. **Cursor.** The checked build is Cursor 3.29.9, observed on 7 November 2026. In an agent shell, `CURSOR_AGENT` was set to a flag (`0`, `1`, `true`, or `false`). A GNOME terminal outside Cursor did not have the name. This check did not include a human terminal opened inside Cursor. Other `CURSOR_*` names in that agent shell are paths, identifiers, or a restore blob, and they are not catalog rows.

3. **Claude Code.** The checked statement is the vendor variables page as read on 7 November 2026, for Claude Code v2.1.172 and later: <https://code.claude.com/docs/en/env-vars>. `CLAUDE_CODE_CHILD_SESSION` is set to `1` on subprocesses Claude Code launches with the Bash, PowerShell, and Monitor tools, and on hook and status-line commands. IDE extensions do not set it. stdio MCP server subprocesses do not receive it. This change did not run a `claude` binary.

4. **Inheritance.** The vendor says a later shell can keep `CLAUDE_CODE_CHILD_SESSION`, including `screen` and a background launcher first started by the Bash tool. As of v2.1.178, Claude Code ignores that inherited value inside tmux. `CURSOR_AGENT` was present in the agent shell and absent from the outside GNOME terminal. Inheritance into a later shell from that Cursor session was not checked.

5. **`CLAUDECODE` is excluded.** The vendor sets it in IDE integrated terminals as well as in Claude Code subprocesses. An ordinary human terminal inside the IDE would match.

6. **Remaining products.** The OpenAI Codex CLI and Gemini CLI stay out of the catalog until a later change records a pinned version for each. Adding, dropping, or replacing a row is a later change to this set.

7. **Opt-in unchanged.** Init still writes policy version 2 and does not write `environment_markers`.

## Consequences

- With `environment_markers = true` and `[agents]` action keys, a matching caller follows those keys ([ADR 0021](0021-policy-schema-v4.md)).
- A shell that inherited `CLAUDE_CODE_CHILD_SESSION` is recognized. That is the inheritance the vendor documents.
- If a human Cursor terminal also sets `CURSOR_AGENT`, a later change drops or replaces that row.
- Codex CLI and Gemini CLI remain open work inside `AG-5`.
