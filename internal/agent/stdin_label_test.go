package agent_test

import (
	"aws-cli-guard/internal/agent"
	"testing"
)

func TestStdinLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		state agent.Stdin
		want  string
	}{
		{agent.StdinInteractive, "interactive"},
		{agent.StdinNonInteractive, "non-interactive"},
		{agent.StdinUndetermined, "undetermined"},
		{agent.Stdin(99), ""},
	}

	for _, testCase := range cases {
		ctx := agent.ContextWithStdin(testCase.state)

		got := ctx.StdinLabel()
		if got != testCase.want {
			t.Fatalf("Stdin(%d).StdinLabel() = %q, want %q", testCase.state, got, testCase.want)
		}
	}
}
