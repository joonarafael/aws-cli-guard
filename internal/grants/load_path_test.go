//nolint:testpackage // tests unexported mapPathError and Corruption path classification.
package grants

import (
	"aws-cli-guard/internal/paths"
	"errors"
	"fmt"
	"testing"
)

func TestMapPathErrorHomeUnset(t *testing.T) {
	t.Parallel()

	err := mapPathError(fmt.Errorf("%w for state", paths.ErrHomeUnset))
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, paths.ErrHomeUnset) {
		t.Fatalf("errors.Is ErrHomeUnset = false, got %v", err)
	}

	var grErr *Error

	if !errors.As(err, &grErr) || grErr.Kind != KindPath {
		t.Fatalf("expected KindPath, got %v", err)
	}
}

func TestCorruptionKindPathFalse(t *testing.T) {
	t.Parallel()

	err := mapPathError(fmt.Errorf("%w for state", paths.ErrHomeUnset))
	if err == nil {
		t.Fatal("expected error")
	}

	_, ok := Corruption(err)
	if ok {
		t.Fatal("Corruption = true for KindPath")
	}
}
