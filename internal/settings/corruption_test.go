//nolint:testpackage // builds synthetic settings.Error values for Corruption classification.
package settings

import (
	"aws-cli-guard/internal/decode"
	"testing"
)

func TestCorruptionInvalidValue(t *testing.T) {
	t.Parallel()

	err := newError(KindInvalid, "/tmp/settings.toml", "identity.verify", 0, 0, nil)

	note, ok := Corruption(err)
	if !ok {
		t.Fatal("Corruption = false, want true")
	}

	if note.Class.Name != "settings" || note.Class.Recover != "aws-guard recover settings" {
		t.Fatalf("note class = %+v", note.Class)
	}
}

func TestCorruptionIgnoresPathAndWrite(t *testing.T) {
	t.Parallel()

	pathErr := newError(KindPath, "", "", 0, 0, nil)

	_, pathCorrupt := Corruption(pathErr)
	if pathCorrupt {
		t.Fatal("KindPath should not be corruption")
	}

	writeErr := newError(KindWrite, "/tmp/settings.toml", "", 0, 0, nil)

	_, writeCorrupt := Corruption(writeErr)
	if writeCorrupt {
		t.Fatal("KindWrite should not be corruption")
	}
}

func TestCorruptionIgnoresProgrammerDecodeSentinel(t *testing.T) {
	t.Parallel()

	err := newError(KindDecode, "/tmp/settings.toml", "", 0, 0, decode.ErrDest)

	_, ok := Corruption(err)
	if ok {
		t.Fatal("Corruption = true, want false for programmer decode sentinel")
	}
}
