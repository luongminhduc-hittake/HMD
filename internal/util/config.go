package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config represents persistent user settings for hmd.
type Config struct {
	DownloadDir         string `json:"download_dir"`
	CookiesBrowser      string `json:"cookies_browser,omitempty"`
	ConcurrentFragments int    `json:"concurrent_fragments,omitempty"`
	MaxRetries          int    `json:"max_retries,omitempty"`
	FragmentRetries     int    `json:"fragment_retries,omitempty"`
}

// AllowedBrowsers contains safe browser identifiers for yt-dlp --cookies-from-browser.
var AllowedBrowsers = map[string]bool{
	"chrome":   true,
	"firefox":  true,
	"brave":    true,
	"edge":     true,
	"chromium": true,
	"opera":    true,
	"vivaldi":  true,
}

// IsAllowedBrowser checks if the given browser name is in the security allowlist.
func IsAllowedBrowser(browser string) bool {
	return AllowedBrowsers[strings.ToLower(strings.TrimSpace(browser))]
}

// CycleBrowser rotates through common browser cookie options: off -> chrome -> firefox -> edge -> brave -> off.
func CycleBrowser(current string) string {
	switch strings.ToLower(strings.TrimSpace(current)) {
	case "chrome":
		return "firefox"
	case "firefox":
		return "edge"
	case "edge":
		return "brave"
	case "brave":
		return ""
	default:
		return "chrome"
	}
}

// GetConfigFilePath returns the path to ~/.config/hmd/config.json.
func GetConfigFilePath() (string, error) {
	if env := os.Getenv("HMD_CONFIG_DIR"); env != "" {
		if err := os.MkdirAll(env, 0755); err != nil {
			return "", err
		}
		return filepath.Join(env, "config.json"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	appConfigDir := filepath.Join(configDir, "hmd")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appConfigDir, "config.json"), nil
}

// LoadConfig loads configuration from disk or returns defaults if absent.
func LoadConfig() Config {
	cfg := Config{
		DownloadDir:         GetDefaultDownloadDir(),
		ConcurrentFragments: 4,
		MaxRetries:          3,
		FragmentRetries:     10,
	}
	cfgPath, err := GetConfigFilePath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if strings.TrimSpace(cfg.DownloadDir) == "" {
		cfg.DownloadDir = GetDefaultDownloadDir()
	}
	if cfg.ConcurrentFragments <= 0 {
		cfg.ConcurrentFragments = 4
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.FragmentRetries <= 0 {
		cfg.FragmentRetries = 10
	}
	return cfg
}

// SaveConfig persists configuration to ~/.config/hmd/config.json.
func SaveConfig(cfg Config) error {
	cfgPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}
