package grants_test

import (
	"aws-cli-guard/internal/grants"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"
)

const (
	overlapGrantOne    = "overlap-one"
	overlapGrantTwo    = "overlap-two"
	revokeProfileDrop  = "drop"
	revokeProfileKeep  = "keep"
	revokeProfileExtra = "extra"
)

func holdUpdateAppend(
	env []string,
	profile string,
	expires time.Time,
	entered chan struct{},
	release chan struct{},
) error {
	err := grants.Update(env, func(current []grants.Grant) ([]grants.Grant, error) {
		close(entered)

		<-release

		out := append(append([]grants.Grant(nil), current...), grants.Grant{
			Profile: profile,
			Expires: expires,
			Source:  emptySource(),
		})

		return out, nil
	})
	if err != nil {
		return fmt.Errorf("hold update append: %w", err)
	}

	return nil
}

func appendGrantUpdate(
	env []string,
	profile string,
	expires time.Time,
	onEnter func([]grants.Grant),
) error {
	err := grants.Update(env, func(current []grants.Grant) ([]grants.Grant, error) {
		onEnter(current)

		out := append(append([]grants.Grant(nil), current...), grants.Grant{
			Profile: profile,
			Expires: expires,
			Source:  emptySource(),
		})

		return out, nil
	})
	if err != nil {
		return fmt.Errorf("append grant update: %w", err)
	}

	return nil
}

func assertMutateNotEntered(t *testing.T, entered chan struct{}, label string) {
	t.Helper()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-entered:
		t.Fatalf("%s entered mutate while first still held lock", label)
	default:
	}
}

func assertOperationNotDone(t *testing.T, done chan struct{}, label string) {
	t.Helper()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
		t.Fatalf("%s finished while Update still held lock", label)
	default:
	}
}

func grantProfiles(list []grants.Grant) []string {
	out := make([]string, len(list))
	for i, grant := range list {
		out[i] = grant.Profile
	}

	return out
}

type dualAppendRun struct {
	secondObserved []string
	firstErr       error
	secondErr      error
}

func startDualAppend(env []string) (*dualAppendRun, chan struct{}, chan struct{}, *sync.WaitGroup) {
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondEntered := make(chan struct{})

	var run dualAppendRun

	var obsMu sync.Mutex

	var waitGroup sync.WaitGroup

	firstExpires := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	secondExpires := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	waitGroup.Go(func() {
		run.firstErr = holdUpdateAppend(
			env,
			overlapGrantOne,
			firstExpires,
			firstEntered,
			releaseFirst,
		)
	})

	<-firstEntered

	waitGroup.Go(func() {
		run.secondErr = appendGrantUpdate(
			env,
			overlapGrantTwo,
			secondExpires,
			func(current []grants.Grant) {
				close(secondEntered)

				obsMu.Lock()
				run.secondObserved = grantProfiles(current)
				obsMu.Unlock()
			},
		)
	})

	return &run, secondEntered, releaseFirst, &waitGroup
}

func finishDualAppend(
	t *testing.T,
	run *dualAppendRun,
	secondEntered chan struct{},
	release chan struct{},
	waitGroup *sync.WaitGroup,
) []string {
	t.Helper()

	assertMutateNotEntered(t, secondEntered, "second Update")

	close(release)

	waitGroup.Wait()

	if run.firstErr != nil {
		t.Fatalf("first Update: %v", run.firstErr)
	}

	if run.secondErr != nil {
		t.Fatalf("second Update: %v", run.secondErr)
	}

	return slices.Clone(run.secondObserved)
}

func seedRevokeOverlapGrants(t *testing.T, env []string) {
	t.Helper()

	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	err := grants.Replace(env, []grants.Grant{
		{Profile: revokeProfileDrop, Expires: base, Source: emptySource()},
		{Profile: revokeProfileKeep, Expires: base, Source: emptySource()},
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
}

func startHeldUpdateExtra(
	t *testing.T,
	env []string,
) (chan struct{}, *sync.WaitGroup, *error) {
	t.Helper()

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})

	var waitGroup sync.WaitGroup

	var errBox error

	extraExpires := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	waitGroup.Go(func() {
		errBox = grants.Update(env, func(current []grants.Grant) ([]grants.Grant, error) {
			close(firstEntered)

			<-releaseFirst

			return append(append([]grants.Grant(nil), current...), grants.Grant{
				Profile: revokeProfileExtra,
				Expires: extraExpires,
				Source:  emptySource(),
			}), nil
		})
	})

	<-firstEntered

	return releaseFirst, &waitGroup, &errBox
}

func finishRevokeOverlap(
	t *testing.T,
	env []string,
	release chan struct{},
	waitGroup *sync.WaitGroup,
	updateErr *error,
) (int, []string) {
	t.Helper()

	revokeDone := make(chan struct{})

	var removed int

	var revokeErr error

	go func() {
		removed, revokeErr = grants.Revoke(env, grants.RevokeRequest{
			All:     false,
			Profile: revokeProfileDrop,
		})

		close(revokeDone)
	}()

	assertOperationNotDone(t, revokeDone, "Revoke")

	close(release)

	waitGroup.Wait()
	<-revokeDone

	if *updateErr != nil {
		t.Fatalf("Update: %v", *updateErr)
	}

	if revokeErr != nil {
		t.Fatalf("Revoke: %v", revokeErr)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return removed, grantProfiles(got.Grants)
}
