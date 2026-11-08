package session_test

import (
	"aws-cli-guard/internal/session"
	"aws-cli-guard/internal/sessions"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	testTTYPath = "/dev/pts/session-off-test"
	testDev     = 5
	testIno     = 99
	testRdev    = 88
	testToken   = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestPolicySkippedLatestOn(t *testing.T) {
	file := sessions.File{ //nolint:exhaustruct_v5 // test fixture partial File.
		Records: []sessions.Record{
			offRecord("sid:1"),
			onRecord("sid:1"),
		},
	}

	stubLstat(t)

	if session.PolicySkipped(file, "sid:1") {
		t.Fatal("want false after on")
	}
}

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestPolicySkippedDifferentSession(t *testing.T) {
	file := sessions.File{ //nolint:exhaustruct_v5 // test fixture partial File.
		Records: []sessions.Record{offRecord("sid:2")},
	}

	stubLstat(t)

	if session.PolicySkipped(file, "sid:1") {
		t.Fatal("want false for other session")
	}
}

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestPolicySkippedMatchingTTYViaLstat(t *testing.T) {
	file := sessions.File{ //nolint:exhaustruct_v5 // test fixture partial File.
		Records: []sessions.Record{offRecord(session.ID())},
	}

	stubLstat(t)

	if !session.PolicySkipped(file, session.ID()) {
		t.Fatal("want skip")
	}
}

//nolint:paralleltest // LstatCharDeviceStats hook.
func TestPolicySkippedMissingDevice(t *testing.T) {
	file := sessions.File{ //nolint:exhaustruct_v5 // test fixture partial File.
		Records: []sessions.Record{offRecord(session.ID())},
	}

	session.LstatCharDeviceStats = func(string) (session.CharDeviceStat, error) {
		return session.CharDeviceStat{}, os.ErrNotExist
	}

	t.Cleanup(func() {
	})

	if session.PolicySkipped(file, session.ID()) {
		t.Fatal("want false when device missing")
	}
}

func stubLstat(t *testing.T) {
	t.Helper()

	prev := session.LstatCharDeviceStats

	session.LstatCharDeviceStats = func(path string) (session.CharDeviceStat, error) {
		if path != testTTYPath {
			return session.CharDeviceStat{}, os.ErrNotExist
		}

		return session.CharDeviceStat{Dev: testDev, Ino: testIno, Rdev: testRdev}, nil
	}

	t.Cleanup(func() {
		session.LstatCharDeviceStats = prev
	})
}

func offRecord(sessionID string) sessions.Record {
	return sessions.Record{
		SessionID: sessionID,
		Effect:    sessions.EffectOff,
		Token:     testToken,
		At:        time.Unix(0, 0).UTC(),
		TTY: &sessions.TTY{
			Path: testTTYPath,
			Dev:  testDev,
			Ino:  testIno,
			Rdev: testRdev,
		},
	}
}

func onRecord(sessionID string) sessions.Record {
	return sessions.Record{
		SessionID: sessionID,
		Effect:    sessions.EffectOn,
		Token:     testToken,
		At:        time.Unix(0, 0).UTC(),
		TTY:       nil,
	}
}

func TestSessionOffWarningNoForgeHints(t *testing.T) {
	t.Parallel()

	if strings.Contains(session.SessionOffWarning, "token") {
		t.Fatal("warning must not mention token")
	}
}
