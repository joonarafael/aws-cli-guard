package gate_test

import (
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/session"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	gateSessionTTYPath = "/dev/pts/gate-session-off"
	gateSessionDev     = 11
	gateSessionIno     = 22
	gateSessionRdev    = 33
	gateSessionToken   = "aaaaaaaabbbbbbbbccccccccddddddddeeeeeeeeffffffff0000000011111111"
)

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestDecideSessionOffHandoffWithWarning(t *testing.T) {
	env := gateEnv(t)
	writeGatePolicy(t, env, denyPolicyBody(gateTestProfile))
	writeGateSessions(t, env, gateSessionOffJSON(session.ID()))

	prev := session.LstatCharDeviceStats
	session.LstatCharDeviceStats = func(path string) (session.CharDeviceStat, error) {
		if path != gateSessionTTYPath {
			return session.CharDeviceStat{}, os.ErrNotExist
		}

		return session.CharDeviceStat{
			Dev:  gateSessionDev,
			Ino:  gateSessionIno,
			Rdev: gateSessionRdev,
		}, nil
	}

	t.Cleanup(func() {
		session.LstatCharDeviceStats = prev
	})

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateSTSService,
			gateGetCallerIdentityOp,
			gateProfileFlag,
			gateTestProfile,
		},
		env,
		"",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	out := stderr.String()
	if !strings.Contains(out, session.SessionOffWarning) {
		t.Fatalf("stderr = %q", out)
	}

	if strings.Contains(out, gateSessionToken) {
		t.Fatal("stderr must not contain session token")
	}
}

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestDecideSessionOffRecognizedAgentStillDenied(t *testing.T) {
	env := gateEnv(t)
	env = append(env, "CURSOR_AGENT=1")
	writeGatePolicy(t, env, productMarkerPolicyPrefix+`

[profiles.`+gateHandoffProfile+`]
action = "allow"

[agents.profiles.`+gateHandoffProfile+`]
action = "deny"
`)
	writeGateSessions(t, env, gateSessionOffJSON(session.ID()))

	prev := session.LstatCharDeviceStats
	session.LstatCharDeviceStats = func(_ string) (session.CharDeviceStat, error) {
		return session.CharDeviceStat{
			Dev: gateSessionDev, Ino: gateSessionIno, Rdev: gateSessionRdev,
		}, nil
	}

	t.Cleanup(func() {
		session.LstatCharDeviceStats = prev
	})

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{
			gateProgName,
			gateProfileFlag,
			gateHandoffProfile,
			gateSTSService,
			gateGetCallerIdentityOp,
		},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false for recognized agent")
	}

	if strings.Contains(stderr.String(), session.SessionOffWarning) {
		t.Fatal("recognized agent must not get session-off warning")
	}
}

func TestDecideCorruptSessionsFile(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGatePolicy(t, env, allowPolicyBody(gateTestProfile))
	writeGateSessions(t, env, `{"version": 1, "records": [`)

	var stderr bytes.Buffer

	permitted := gate.Decide(
		[]string{gateProgName, gateSTSService, gateGetCallerIdentityOp},
		env,
		"",
		&stderr,
	)
	if permitted {
		t.Fatal("Decide = true, want false")
	}

	assertGateCancel(t, stderr.String())

	if !strings.Contains(stderr.String(), "aws-guard recover sessions") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDecideSSOLoginIgnoresCorruptSessions(t *testing.T) {
	t.Parallel()

	env := gateEnv(t)
	writeGateSessions(t, env, `not json`)

	var stderr bytes.Buffer

	permitted := gate.Decide(gateSSOLoginArgs(), env, "", &stderr)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func writeGateSessions(t *testing.T, env []string, body string) {
	t.Helper()

	dir := stateDirFromEnv(t, env)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "sessions.json"), []byte(body), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func gateSessionOffJSON(sessionID string) string {
	return `{
  "version": 1,
  "records": [
    {
      "session_id": "` + sessionID + `",
      "effect": "off",
      "token": "` + gateSessionToken + `",
      "at": "2026-01-01T00:00:00Z",
      "tty": {
        "path": "` + gateSessionTTYPath + `",
        "dev": ` + gateSessionInt(gateSessionDev) + `,
        "ino": ` + gateSessionInt(gateSessionIno) + `,
        "rdev": ` + gateSessionInt(gateSessionRdev) + `
      }
    }
  ]
}
`
}

func gateSessionInt(v int) string {
	return strconv.Itoa(v)
}
