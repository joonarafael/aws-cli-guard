//nolint:testpackage // exercises unexported matchMarkers and recognition.
package agent

import (
	"testing"
)

const testAgentMarkerEnv = "AWS_CLI_GUARD_TEST_MARKER"

func TestMatchMarkersNonEmptyValue(t *testing.T) {
	t.Parallel()

	env := []string{testAgentMarkerEnv + "=set"}

	if !matchMarkers(env, []string{testAgentMarkerEnv}) {
		t.Fatal("want match")
	}
}

func TestMatchMarkersEmptyValueAndMissing(t *testing.T) {
	t.Parallel()

	if matchMarkers([]string{testAgentMarkerEnv + "="}, []string{testAgentMarkerEnv}) {
		t.Fatal("empty value should not match")
	}

	if matchMarkers([]string{"OTHER=1"}, []string{testAgentMarkerEnv}) {
		t.Fatal("missing name should not match")
	}
}

func TestMatchMarkersFirstWinEmptyDoesNotMatch(t *testing.T) {
	t.Parallel()

	env := []string{
		testAgentMarkerEnv + "=",
		testAgentMarkerEnv + "=set",
	}

	if matchMarkers(env, []string{testAgentMarkerEnv}) {
		t.Fatal("first-wins empty should not match")
	}
}

func TestMatchMarkersTwoNamesStillOneMatch(t *testing.T) {
	t.Parallel()

	env := []string{"MARKER_A=1", "MARKER_B=1"}

	if !matchMarkers(env, []string{"MARKER_A", "MARKER_B"}) {
		t.Fatal("want match")
	}
}

func TestApplyMarkersDisabledLeavesUnknown(t *testing.T) {
	t.Parallel()

	ctx := ContextWithStdin(StdinInteractive)
	ctx = ApplyMarkers(ctx, []string{testAgentMarkerEnv + "=1"}, false)

	if ctx.Recognized() {
		t.Fatal("want unknown when disabled")
	}
}

func TestApplyMarkersBuiltinCatalogEmpty(t *testing.T) {
	t.Parallel()

	ctx := ContextWithStdin(StdinInteractive)
	ctx = ApplyMarkers(ctx, []string{"ANY_VAR=1"}, true)

	if ctx.Recognized() {
		t.Fatal("empty builtin catalog should not recognize")
	}
}

func TestApplyMarkersRecognitionKeepsStdinClass(t *testing.T) {
	t.Parallel()

	ctx := ContextWithStdin(StdinNonInteractive)
	_ = matchMarkers([]string{"X=1"}, []string{"X"})

	ctx.rec = recognitionAgent

	if !ctx.Recognized() || !ctx.FollowsProfilePolicy() {
		t.Fatal("recognized caller should still follow profile policy")
	}
}
