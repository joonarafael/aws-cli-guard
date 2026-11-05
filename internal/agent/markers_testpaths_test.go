//go:build awsguard_testpaths

package agent_test

import (
	"aws-cli-guard/internal/agent"
	"testing"
)

func TestApplyMarkersRecognizesTestAgentMarker(t *testing.T) {
	t.Parallel()

	ctx := agent.ContextWithStdin(agent.StdinInteractive)
	ctx = agent.ApplyMarkers(ctx, []string{agent.TestAgentMarkerEnv + "=1"}, true)

	if !ctx.Recognized() {
		t.Fatal("want recognized with test marker")
	}
}
