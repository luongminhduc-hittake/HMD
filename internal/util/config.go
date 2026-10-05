package util

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents persistent user settings for ytdl.
type Config struct {
	DownloadDir string `json:"download_dir"`
}

// GetConfigFilePath returns the path to ~/.config/ytdl/config.json.
func GetConfigFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	appConfigDir := filepath.Join(configDir, "ytdl")
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

// SaveConfig persists configuration to ~/.config/ytdl/config.json.
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

// HasNativePicker checks if a native GUI directory picker tool is available.
func HasNativePicker() bool {
	if runtime.GOOS == "windows" {
		_, err := exec.LookPath("powershell")
		return err == nil
	}
	if _, err := exec.LookPath("zenity"); err == nil {
		return true
	}
	if _, err := exec.LookPath("kdialog"); err == nil {
		return true
	}
	return false
}

// PickFolderNative triggers the OS native folder chooser dialog.
// Returns (path, isCancelled, error).
func PickFolderNative() (string, bool, error) {
	if runtime.GOOS == "windows" {
		psScript := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = 'Chọn thư mục lưu tải về YouTube'; if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $f.SelectedPath }`
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
		out, err := cmd.Output()
		if err != nil {
			return "", false, err
		}
		res := strings.TrimSpace(string(out))
		if res == "" {
			return "", true, nil
		}
		return res, false, nil
	}

	if path, err := exec.LookPath("zenity"); err == nil {
		cmd := exec.Command(path, "--file-selection", "--directory", "--title=Chọn thư mục lưu YouTube")
		out, err := cmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				// User clicked Cancel or closed dialog
				return "", true, nil
			}
			return "", false, err
		}
		res := strings.TrimSpace(string(out))
		if res == "" {
			return "", true, nil
		}
		return res, false, nil
	}

	if path, err := exec.LookPath("kdialog"); err == nil {
		cmd := exec.Command(path, "--getexistingdirectory")
		out, err := cmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				return "", true, nil
			}
			return "", false, err
		}
		res := strings.TrimSpace(string(out))
		if res == "" {
			return "", true, nil
		}
		return res, false, nil
	}

	return "", false, errors.New("no native picker available")
}
