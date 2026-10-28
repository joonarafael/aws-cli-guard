package decode_test

import (
	"aws-cli-guard/internal/decode"
	"strings"
	"testing"
)

func TestErrorsDoNotLeakSecrets(t *testing.T) {
	t.Parallel()

	secret := fixtureSecret

	for _, testCase := range secretLeakCases(secret) {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.run()
			assertErrorOmitsSecret(t, err, secret)
		})
	}
}

func secretLeakCases(secret string) []struct {
	name string
	run  func() error
} {
	return []struct {
		name string
		run  func() error
	}{
		{
			name: "toml-syntax",
			run: func() error {
				var doc tomlDoc

				data := []byte("version = \"" + secret + "\"\n")

				return decode.Decode(decode.TOML, data, &doc, []int{1})
			},
		},
		{
			name: "toml-duplicate",
			run: func() error {
				var doc tomlDoc

				data := []byte("name = \"" + secret + "\"\nname = \"dup\"\nversion = 1\n")

				return decode.Decode(decode.TOML, data, &doc, []int{1})
			},
		},
		{
			name: "toml-unknown",
			run: func() error {
				var doc tomlDoc

				data := []byte("version = 1\nextra = \"" + secret + "\"\n")

				return decode.Decode(decode.TOML, data, &doc, []int{1})
			},
		},
		{
			name: "json-unknown",
			run: func() error {
				var doc jsonDoc

				data := []byte(`{"version":1,"extra":"` + secret + `"}`)

				return decode.Decode(decode.JSON, data, &doc, []int{1})
			},
		},
	}
}

func assertErrorOmitsSecret(t *testing.T, err error, secret string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}

	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaks secret: %v", err)
	}
}
