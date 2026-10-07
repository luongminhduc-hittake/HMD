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
	"auto":      true,
	"chrome":    true,
	"firefox":   true,
	"brave":     true,
	"edge":      true,
	"chromium":  true,
	"opera":     true,
	"vivaldi":   true,
	"zen":       true,
	"floorp":    true,
	"librewolf": true,
	"waterfox":  true,
	"coccoc":    true,
}

// IsAllowedBrowser checks if the given browser name is in the security allowlist
// or matches safe custom profile syntax like "firefox:/path".
func IsAllowedBrowser(browser string) bool {
	b := strings.TrimSpace(browser)
	if b == "" {
		return false
	}
	if AllowedBrowsers[strings.ToLower(b)] {
		return true
	}
	for _, prefix := range []string{"firefox:", "chromium:"} {
		if strings.HasPrefix(b, prefix) {
			targetPath := strings.TrimPrefix(b, prefix)
			if strings.ContainsAny(targetPath, ";|&$`\r\n\t") || len(targetPath) == 0 {
				return false
			}
			return true
		}
	}
	return false
}

// CycleBrowser rotates through available browser cookie options:
// off ("") -> "auto" -> [detected installed browsers...] -> off ("")
func CycleBrowser(current string) string {
	current = strings.ToLower(strings.TrimSpace(current))
	options := []string{"", "auto"}

	installed := ScanInstalledBrowsers()
	if len(installed) > 0 {
		seen := make(map[string]bool)
		for _, b := range installed {
			if !seen[b.ID] {
				seen[b.ID] = true
				options = append(options, b.ID)
			}
		}
	} else {
		options = append(options, "chrome", "firefox", "edge", "brave")
	}

	for i, opt := range options {
		if opt == current {
			return options[(i+1)%len(options)]
		}
	}
	return "auto"
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
