package decode_test

import (
	"aws-cli-guard/internal/decode"
	"errors"
	"testing"
)

const fixtureSecret = "AKIAIOSFODNN7EXAMPLE-secret-body"

type tomlDoc struct {
	Version int    `toml:"version"`
	Name    string `toml:"name"`
}

type jsonDoc struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
}

type jsonNested struct {
	Version int `json:"version"`
	Inner   struct {
		Flag bool `json:"flag"`
	} `json:"inner"`
}

func TestDecodeTOMLSuccess(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 1\nname = \"dev\"\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if doc.Version != 1 || doc.Name != "dev" {
		t.Fatalf("got %+v", doc)
	}
}

func TestDecodeJSONSuccess(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"version":1,"name":"dev"}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if doc.Version != 1 || doc.Name != "dev" {
		t.Fatalf("got %+v", doc)
	}
}

func TestDecodeAcceptsOlderSupportedVersion(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 1\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1, 2})
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if doc.Version != 1 {
		t.Fatalf("version = %d", doc.Version)
	}
}

func TestDecodeProgrammerErrors(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	cases := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "format",
			run: func() error {
				return decode.Decode(decode.Format(0), []byte("version = 1\n"), &doc, []int{1})
			},
			want: decode.ErrFormat,
		},
		{
			name: "supported",
			run: func() error {
				return decode.Decode(decode.TOML, []byte("version = 1\n"), &doc, nil)
			},
			want: decode.ErrSupported,
		},
		{
			name: "dest",
			run: func() error {
				return decode.Decode(decode.TOML, []byte("version = 1\n"), nil, []int{1})
			},
			want: decode.ErrDest,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.run()
			if !errors.Is(err, testCase.want) {
				t.Fatalf("errors.Is(%v, %v) = false", err, testCase.want)
			}
		})
	}
}

func TestDecodeTOMLUnsupportedBeforeUnknownField(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 99\nname = \"x\"\nextra = true\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindUnsupportedVersion)
}

func TestDecodeJSONUnsupportedBeforeUnknownField(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"version":99,"name":"x","extra":true}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindUnsupportedVersion)
}

func TestDecodeTOMLDuplicateBeforeBadVersion(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 99\nversion = 1\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindDuplicateKey)
}

func TestDecodeJSONDuplicateBeforeBadVersion(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"version":99,"version":1}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindDuplicateKey)
}

func TestDecodeJSONNestedDuplicate(t *testing.T) {
	t.Parallel()

	var doc jsonNested

	data := []byte(`{"version":1,"inner":{"flag":true,"flag":false}}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindDuplicateKey)
}

func TestDecodeExplicitZeroVersionUnsupported(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 0\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindUnsupportedVersion)
}

func TestDecodeMissingVersion(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"name":"only"}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindMissingVersion)
}

func TestDecodeVersionTypeJSONFloat(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"version":1.0,"name":"x"}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindVersionType)
}

func TestDecodeUnknownFieldTOML(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 1\nextra = true\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindUnknownField)
}

func TestDecodeUnknownFieldJSON(t *testing.T) {
	t.Parallel()

	var doc jsonDoc

	data := []byte(`{"version":1,"extra":true}`)

	err := decode.Decode(decode.JSON, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindUnknownField)
}

func TestDecodeTypeMismatchTOML(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	data := []byte("version = 1\nname = 42\n")

	err := decode.Decode(decode.TOML, data, &doc, []int{1})
	assertDocKind(t, err, decode.KindType)
}

func TestDecodeSyntaxEmpty(t *testing.T) {
	t.Parallel()

	var doc tomlDoc

	err := decode.Decode(decode.TOML, []byte("   \n"), &doc, []int{1})
	assertDocKind(t, err, decode.KindSyntax)
}

func assertDocKind(t *testing.T, err error, kind decode.Kind) {
	t.Helper()

	var docErr *decode.Error
	if !errors.As(err, &docErr) {
		t.Fatalf("expected *decode.Error, got %T: %v", err, err)
	}

	if docErr.Kind != kind {
		t.Fatalf("Kind = %v, want %v (err %v)", docErr.Kind, kind, err)
	}
}
