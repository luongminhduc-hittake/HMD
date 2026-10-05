package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config represents persistent user settings for hmd.
type Config struct {
	DownloadDir string `json:"download_dir"`
}

// GetConfigFilePath returns the path to ~/.config/hmd/config.json.
func GetConfigFilePath() (string, error) {
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
		DownloadDir: GetDefaultDownloadDir(),
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
