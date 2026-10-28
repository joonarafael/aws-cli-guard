package grants_test

import (
	"aws-cli-guard/internal/grants"
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReplaceRoundTrip(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	expires := time.Date(2026, 10, 28, 18, 0, 0, 0, time.UTC)
	input := []grants.Grant{
		{
			Profile: "production-admin",
			Expires: expires,
			Source: grants.Source{
				AccountID:         "123456789012",
				RoleARN:           "arn:aws:iam::123456789012:role/Admin",
				SSOAccountID:      "",
				SSORoleName:       "",
				CredentialProcess: false,
			},
		},
	}

	err := grants.Replace(env, input)
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 1 || !got.Grants[0].Expires.Equal(expires) {
		t.Fatalf("got %+v", got.Grants)
	}

	info, err := os.Stat(grantsPathFromEnv(t, env))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}

	dirInfo, err := os.Stat(stateDirFromEnv(t, env))
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}

	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %o", dirInfo.Mode().Perm())
	}
}

func TestReplaceNilEmptyFile(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	err := grants.Replace(env, nil)
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 0 {
		t.Fatalf("grants = %+v", got.Grants)
	}

	data, err := os.ReadFile(grantsPathFromEnv(t, env))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := `{
  "version": 1,
  "grants": []
}
`
	if !bytes.Equal(data, []byte(want)) {
		t.Fatalf("body:\n%s", data)
	}
}

func TestReplaceEncodesOffsetAsZ(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	local := time.Date(2026, 6, 15, 12, 30, 45, 0, time.FixedZone("CET", 2*3600))

	err := grants.Replace(env, []grants.Grant{
		{
			Profile: "p",
			Expires: local,
			Source:  emptySource(),
		},
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	data, err := os.ReadFile(grantsPathFromEnv(t, env))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !bytes.Contains(data, []byte(`"expires": "2026-06-15T10:30:45Z"`)) {
		t.Fatalf("body:\n%s", data)
	}
}

func TestUpdateAppendsGrant(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	second := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	err := grants.Replace(env, []grants.Grant{
		{Profile: "one", Expires: first, Source: emptySource()},
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	err = grants.Update(env, func(current []grants.Grant) ([]grants.Grant, error) {
		out := append(append([]grants.Grant(nil), current...), grants.Grant{
			Profile: "two",
			Expires: second,
			Source:  emptySource(),
		})

		return out, nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Grants) != 2 {
		t.Fatalf("len = %d", len(got.Grants))
	}

	if got.Grants[0].Profile != "one" || got.Grants[1].Profile != "two" {
		t.Fatalf("grants = %+v", got.Grants)
	}
}

func TestUpdateOnCorruptFileUnchanged(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	path := grantsPathFromEnv(t, env)
	corrupt := []byte(`{not json`)

	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile(path, corrupt, 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err = grants.Update(env, func(_ []grants.Grant) ([]grants.Grant, error) {
		return []grants.Grant{}, nil
	})
	if err == nil {
		t.Fatal("expected error")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !bytes.Equal(after, corrupt) {
		t.Fatal("corrupt file was modified")
	}
}
