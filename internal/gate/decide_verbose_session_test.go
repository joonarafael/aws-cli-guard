package gate_test

import (
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/session"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDecideSessionOffVerboseNoTokenOrPath(t *testing.T) {
	t.Parallel()

	env := append(gateEnv(t), "AWS_CLI_GUARD_LOG=verbose")
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
		"/bin/aws",
		&stderr,
	)
	if !permitted {
		t.Fatalf("Decide = false, stderr = %q", stderr.String())
	}

	out := stderr.String()
	if !strings.Contains(out, session.SessionOffWarning) {
		t.Fatalf("stderr = %q", out)
	}

	if !strings.Contains(out, "outcome: session off") {
		t.Fatalf("stderr = %q", out)
	}

	if strings.Contains(out, "token") || strings.Contains(out, "sessions.json") {
		t.Fatalf("stderr leaked session internals: %q", out)
	}
}
