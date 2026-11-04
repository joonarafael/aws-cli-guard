package grants_test

import (
	"aws-cli-guard/internal/grants"
	"slices"
	"testing"
)

func TestUpdateConcurrentAppendsBothGrants(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	run, secondEntered, release, waitGroup := startDualAppend(env)
	observed := finishDualAppend(t, run, secondEntered, release, waitGroup)

	if !slices.Contains(observed, overlapGrantOne) {
		t.Fatalf("second Update saw profiles %v, want %q", observed, overlapGrantOne)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	profiles := grantProfiles(got.Grants)

	if !slices.Contains(profiles, overlapGrantOne) || !slices.Contains(profiles, overlapGrantTwo) {
		t.Fatalf("grants profiles = %v, want %q and %q", profiles, overlapGrantOne, overlapGrantTwo)
	}
}

func TestRevokeSeesHeldUpdate(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	seedRevokeOverlapGrants(t, env)

	release, waitGroup, updateErr := startHeldUpdateExtra(t, env)
	removed, profiles := finishRevokeOverlap(t, env, release, waitGroup, updateErr)

	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	if slices.Contains(profiles, revokeProfileDrop) {
		t.Fatalf("grants still contain %q: %v", revokeProfileDrop, profiles)
	}

	hasKeep := slices.Contains(profiles, revokeProfileKeep)

	hasExtra := slices.Contains(profiles, revokeProfileExtra)
	if !hasKeep || !hasExtra {
		t.Fatalf(
			"grants profiles = %v, want %q and %q",
			profiles,
			revokeProfileKeep,
			revokeProfileExtra,
		)
	}
}
