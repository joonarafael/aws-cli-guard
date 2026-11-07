package agent

// claudeCodeChildSessionEnv is the Claude Code marker recorded in ADR 0022.
const claudeCodeChildSessionEnv = "CLAUDE_CODE_CHILD_SESSION"

// cursorAgentEnv is the Cursor marker recorded in ADR 0022.
const cursorAgentEnv = "CURSOR_AGENT"

func productionBuiltinMarkerNames() []string {
	return []string{
		claudeCodeChildSessionEnv,
		cursorAgentEnv,
	}
}
