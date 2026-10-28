package awsconfig_test

import (
	"aws-cli-guard/internal/awsconfig"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFilesEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{"HOME=" + dir}

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Profiles) != 0 {
		t.Fatalf("profiles = %+v", cfg.Profiles)
	}
}

func TestLoadConfigOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "custom-config")
	writeFile(t, configPath, "[profile x]\nregion = sa-east-1\n")

	env := []string{
		"HOME=" + dir,
		"AWS_CONFIG_FILE=" + configPath,
	}

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Profiles["x"].Settings["region"] != "sa-east-1" {
		t.Fatalf("profile x: %+v", cfg.Profiles["x"])
	}
}

func TestLoadDirectoryPathErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	env := []string{
		"HOME=" + dir,
		"AWS_CONFIG_FILE=" + dir,
	}

	_, err := awsconfig.Load(env)
	if err == nil {
		t.Fatal("expected read error for directory path")
	}

	if !errors.Is(err, awsconfig.ErrNotRegular) {
		t.Fatalf("errors.Is ErrNotRegular = false, got %v", err)
	}
}

func TestLoadSymlinkToFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "real-config")
	writeFile(t, target, "[default]\nregion = ap-south-1\n")

	link := filepath.Join(dir, "link-config")

	err := os.Symlink(target, link)
	if err != nil {
		t.Fatalf("symlink: %v", err)
	}

	env := []string{
		"HOME=" + dir,
		"AWS_CONFIG_FILE=" + link,
	}

	cfg, err := awsconfig.Load(env)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Profiles["default"].Settings["region"] != "ap-south-1" {
		t.Fatalf("default: %+v", cfg.Profiles["default"])
	}
}
