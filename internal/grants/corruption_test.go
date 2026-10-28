package grants_test

import (
	"aws-cli-guard/internal/diagnose"
	"aws-cli-guard/internal/grants"
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestCorruptionDecodeNote(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{"version":1,"grants":[],"extra":true}`+"\n")

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	note, ok := grants.Corruption(err)
	if !ok {
		t.Fatal("Corruption = false")
	}

	if note.Class.Name != "grants" {
		t.Fatalf("class = %q", note.Class.Name)
	}

	var buf bytes.Buffer

	diagnose.Write(&buf, note)

	out := buf.String()
	if !strings.Contains(out, "aws-guard recover grants") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestCorruptionPathErrorFalse(t *testing.T) {
	t.Parallel()

	_, err := grants.Load(nil)
	if err == nil {
		t.Fatal("expected error")
	}

	_, ok := grants.Corruption(err)
	if ok {
		t.Fatal("Corruption = true for path error")
	}
}

func TestCorruptionInvalidValue(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrants(t, env, `{"version":1}`+"\n")

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	note, ok := grants.Corruption(err)
	if !ok || note.Failure != "invalid value" {
		t.Fatalf("note = %+v ok=%v err=%v", note, ok, err)
	}
}

func TestCorruptionTooLarge(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)

	body := make([]byte, (1<<20)+1)
	for i := range body {
		body[i] = ' '
	}

	copy(body, []byte(`{"version":1,"grants":[]}`))
	writeGrantsPerm(t, env, string(body), 0o600)

	_, err := grants.Load(env)
	if err == nil {
		t.Fatal("expected error")
	}

	note, ok := grants.Corruption(err)
	if !ok || note.Failure != "file too large" {
		t.Fatalf("note = %+v", note)
	}
}

func TestLooseModeWarning(t *testing.T) {
	t.Parallel()

	env := stateDirEnv(t)
	writeGrantsPerm(t, env, `{"version":1,"grants":[]}`+"\n", 0o644)

	got, err := grants.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var buf bytes.Buffer

	diagnose.WriteLooseMode(&buf, diagnose.ClassGrants(), got.Path, got.Mode)

	if !strings.Contains(buf.String(), "loose permissions") {
		t.Fatalf("output:\n%s", buf.String())
	}
}

func TestCorruptionWriteErrorFalse(t *testing.T) {
	t.Parallel()

	//nolint:err113 // deliberate non-grants error to test classification
	plain := errors.New("plain")

	_, ok := grants.Corruption(plain)
	if ok {
		t.Fatal("Corruption = true for non-grants error")
	}
}
