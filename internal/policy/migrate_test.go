package policy_test

import (
	"aws-cli-guard/internal/policy"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMigrateVersion1To4PreservesActions(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = "deny"

[profiles.prod]
action = "prompt"
approval_ttl = "5m"
`)

	now := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)

	result, err := policy.Migrate(env, now)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if !result.Rewrote || result.FromVersion != 1 || result.Aside == "" {
		t.Fatalf("result: %+v", result)
	}

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 4 || got.DefaultAction != policy.ActionDeny {
		t.Fatalf("policy: %+v", got)
	}

	if got.HasEnvironmentAction {
		t.Fatal("unexpected environment section")
	}

	prof, ok := got.Profiles["prod"]
	if !ok || prof.Action != policy.ActionPrompt || prof.ApprovalTTL != 5*time.Minute {
		t.Fatalf("profile: %+v", prof)
	}

	body := mustReadPolicy(t, env)
	if strings.Contains(body, "user_declared") || strings.Contains(body, "environment_markers") {
		t.Fatalf("body should not enable detection: %s", body)
	}
}

func TestMigrateVersion3MarkersFalse(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, v3PolicyPrefix+`
[agents.detection]
environment_markers = false
`)

	_, err := policy.Migrate(env, time.Now())
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	got, err := policy.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Version != 4 || !got.HasEnvironmentMarkers || got.EnvironmentMarkers {
		t.Fatalf("markers: %+v", got)
	}
}

func TestMigrateVersion4LeavesFileUnchanged(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	original := `# keep
version = 4

[defaults]
action = "allow"
`
	writePolicy(t, env, original)

	before := mustReadPolicy(t, env)

	_, err := policy.Migrate(env, time.Now())
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	after := mustReadPolicy(t, env)
	if after != before {
		t.Fatalf("file changed:\nbefore %q\nafter %q", before, after)
	}
}

func TestMigrateTwiceOneAside(t *testing.T) {
	t.Parallel()

	env := configDirEnv(t)
	writePolicy(t, env, `version = 1

[defaults]
action = "prompt"
`)

	now := time.Now()

	first, err := policy.Migrate(env, now)
	if err != nil || !first.Rewrote {
		t.Fatalf("first: %+v err %v", first, err)
	}

	second, err := policy.Migrate(env, now)
	if err != nil || second.Rewrote || second.Aside != "" {
		t.Fatalf("second: %+v err %v", second, err)
	}
}

func mustReadPolicy(t *testing.T, env []string) string {
	t.Helper()

	dir := policyDirFromEnv(t, env)
	path := filepath.Join(dir, "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	return string(data)
}
