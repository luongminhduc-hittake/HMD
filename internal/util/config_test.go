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

	// Save custom dir and cookies
	customDir := filepath.Join(tmpDir, "custom_dl")
	cfg.DownloadDir = customDir
	cfg.CookiesBrowser = "chrome"
	cfg.ConcurrentFragments = 6
	cfg.MaxRetries = 5
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Reload
	reloaded := LoadConfig()
	if reloaded.DownloadDir != customDir {
		t.Errorf("expected %q, got %q", customDir, reloaded.DownloadDir)
	}
	if reloaded.CookiesBrowser != "chrome" {
		t.Errorf("expected cookies_browser chrome, got %q", reloaded.CookiesBrowser)
	}
	if reloaded.ConcurrentFragments != 6 {
		t.Errorf("expected concurrent_fragments 6, got %d", reloaded.ConcurrentFragments)
	}
	if reloaded.MaxRetries != 5 {
		t.Errorf("expected max_retries 5, got %d", reloaded.MaxRetries)
	}
}

func TestBrowserCycleAndValidation(t *testing.T) {
	if !IsAllowedBrowser("chrome") {
		t.Errorf("expected chrome to be allowed")
	}
	if !IsAllowedBrowser("FireFox") {
		t.Errorf("expected firefox case-insensitive to be allowed")
	}
	if !IsAllowedBrowser("auto") {
		t.Errorf("expected auto to be allowed")
	}
	if !IsAllowedBrowser("zen") {
		t.Errorf("expected zen to be allowed")
	}
	if !IsAllowedBrowser("firefox:/home/user/.zen/profile") {
		t.Errorf("expected custom firefox profile path to be allowed")
	}
	if IsAllowedBrowser("firefox:/path; rm -rf") {
		t.Errorf("unsafe profile path with semicolon should not be allowed")
	}
	if IsAllowedBrowser("malicious_browser; rm -rf") {
		t.Errorf("unsafe browser should not be allowed")
	}

	// Test rotation starts with auto
	b := CycleBrowser("")
	if b != "auto" {
		t.Errorf("expected cycle from empty to auto, got %s", b)
	}

	// Cycling eventually reaches back to empty string
	visited := map[string]bool{b: true}
	for i := 0; i < 20; i++ {
		b = CycleBrowser(b)
		if b == "" {
			break
		}
		visited[b] = true
	}
	if b != "" {
		t.Errorf("expected cycling to eventually turn off, ended with %s", b)
	}
}

func TestHMDConfigDirEnv(t *testing.T) {
	customDir := t.TempDir()
	t.Setenv("HMD_CONFIG_DIR", customDir)

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(customDir, "config.json")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

