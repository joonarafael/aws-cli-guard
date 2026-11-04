//nolint:testpackage // tests unexported callerAllowsProfile helper.
package gate

import (
	"aws-cli-guard/internal/agent"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCallerAllowsProfileKnownClasses(t *testing.T) {
	t.Parallel()

	for _, state := range []agent.Stdin{
		agent.StdinUndetermined,
		agent.StdinInteractive,
		agent.StdinNonInteractive,
	} {
		t.Run(fmt.Sprintf("stdin_%d", state), func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer

			if !callerAllowsProfile(agent.ContextWithStdin(state), &stderr) {
				t.Fatal("want true")
			}

			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q", stderr.String())
			}
		})
	}
}

func TestCallerAllowsProfileOutOfRange(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	if callerAllowsProfile(agent.ContextWithStdin(agent.Stdin(3)), &stderr) {
		t.Fatal("want false")
	}

	if !strings.Contains(stderr.String(), "aws-cli-guard: error: command was not executed") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	if !strings.Contains(stderr.String(), reasonCallerUnclassified) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
