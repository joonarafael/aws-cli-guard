//go:build !awsguard_testpaths

package agent_test

import (
	"aws-cli-guard/internal/agent"
	"testing"
)

func TestApplyMarkersIgnoresTestAgentMarkerInProductionBuild(t *testing.T) {
	t.Parallel()

	ctx := agent.ContextWithStdin(agent.StdinInteractive)
	ctx = agent.ApplyMarkers(ctx, []string{"AWS_CLI_GUARD_TEST_AGENT=1"}, true)

	if ctx.Recognized() {
		t.Fatal("production build must not recognize test marker")
	}
}
