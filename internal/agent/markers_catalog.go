package agent

// claudeCodeChildSessionEnv is vendor-documented for Claude Code v2.1.172 and later;
// IDE terminals do not set it.
const claudeCodeChildSessionEnv = "CLAUDE_CODE_CHILD_SESSION"

// cursorAgentEnv was observed on Cursor 3.29.9.
const cursorAgentEnv = "CURSOR_AGENT"

func productionBuiltinMarkerNames() []string {
	return []string{
		claudeCodeChildSessionEnv,
		cursorAgentEnv,
	}
}
