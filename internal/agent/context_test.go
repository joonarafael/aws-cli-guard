//nolint:testpackage // exercises unexported recognition and stdin classification fields.
package agent

import "testing"

func TestFollowsProfilePolicyKnownClasses(t *testing.T) {
	t.Parallel()

	for _, state := range []Stdin{StdinUndetermined, StdinInteractive, StdinNonInteractive} {
		ctx := ContextWithStdin(state)
		if !ctx.FollowsProfilePolicy() {
			t.Fatalf("Stdin(%d): want true", state)
		}
	}
}

func TestFollowsProfilePolicyOutOfRange(t *testing.T) {
	t.Parallel()

	ctx := ContextWithStdin(Stdin(3))
	if ctx.FollowsProfilePolicy() {
		t.Fatal("want false")
	}
}

func TestObserveNotRecognized(t *testing.T) {
	t.Parallel()

	ctx := Observe(nil)
	if ctx.Recognized() {
		t.Fatal("Observe(nil) recognized")
	}

	ctx = ContextWithStdin(StdinInteractive)
	if ctx.Recognized() {
		t.Fatal("ContextWithStdin recognized")
	}
}

func TestForcedRecognitionStillFollowsProfilePolicy(t *testing.T) {
	t.Parallel()

	ctx := ContextWithStdin(StdinInteractive)
	ctx.rec = recognition(1)

	if !ctx.Recognized() {
		t.Fatal("want recognized")
	}

	if !ctx.FollowsProfilePolicy() {
		t.Fatal("recognized caller should still follow profile policy in AG-1")
	}
}
