//nolint:testpackage // builds synthetic policy.Error values for Corruption classification.
package policy

import (
	"aws-cli-guard/internal/decode"
	"testing"
)

func TestCorruptionIgnoresProgrammerDecodeSentinel(t *testing.T) {
	t.Parallel()

	err := newError(KindDecode, "/tmp/config.toml", "", 0, 0, decode.ErrDest)

	_, ok := Corruption(err)
	if ok {
		t.Fatal("Corruption = true, want false for programmer decode sentinel")
	}
}

func TestCorruptionIgnoresDecodeWithoutDocError(t *testing.T) {
	t.Parallel()

	err := newError(KindDecode, "/tmp/config.toml", "", 0, 0, decode.ErrFormat)

	_, ok := Corruption(err)
	if ok {
		t.Fatal("Corruption = true, want false when unwrap is not *decode.Error")
	}
}
