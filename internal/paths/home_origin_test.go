//nolint:testpackage // exercises unexported home resolution with injected lookups.
package paths

import (
	"testing"
)

func TestHomeOriginUserDatabase(t *testing.T) {
	t.Parallel()

	env := []string{ignoredHomeEnvEntry}

	label, ok := homeOriginWithLookup(env, lookupReturning("/db/home"))
	if !ok {
		t.Fatal("expected ok")
	}

	if label != homeOriginUserDatabase {
		t.Fatalf("label = %q", label)
	}
}

func TestHomeOriginHOME(t *testing.T) {
	t.Parallel()

	env := []string{fallbackHomeEnvEntry}

	label, ok := homeOriginWithLookup(env, lookupFailing())
	if !ok {
		t.Fatal("expected ok")
	}

	if label != homeOriginHOME {
		t.Fatalf("label = %q", label)
	}
}

func TestHomeOriginUnset(t *testing.T) {
	t.Parallel()

	_, ok := homeOriginWithLookup(nil, lookupFailing())
	if ok {
		t.Fatal("expected not ok")
	}
}

func homeOriginWithLookup(env []string, lookup homeLookup) (string, bool) {
	if dirOverrideActive(env) {
		return "", false
	}

	res := resolveHome(env, lookup)
	if !res.ok {
		return "", false
	}

	return res.origin, true
}
