//go:build awsguard_testpaths

package agent

// TestAgentMarkerEnv is a built-in marker name in test builds only.
const TestAgentMarkerEnv = "AWS_CLI_GUARD_TEST_AGENT"

func extraBuiltinMarkerNames() []string {
	return []string{TestAgentMarkerEnv}
}
