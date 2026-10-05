package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)

	// Defaults when absent
	cfg := LoadConfig()
	if cfg.DownloadDir != GetDefaultDownloadDir() {
		t.Errorf("expected default download dir %q, got %q", GetDefaultDownloadDir(), cfg.DownloadDir)
	}

	// Save custom dir
	customDir := filepath.Join(tmpDir, "custom_dl")
	cfg.DownloadDir = customDir
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Reload
	reloaded := LoadConfig()
	if reloaded.DownloadDir != customDir {
		t.Errorf("expected %q, got %q", customDir, reloaded.DownloadDir)
	}
}
