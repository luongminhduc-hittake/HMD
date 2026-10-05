package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetAppDir returns the directory where hittakeMD stores configuration and binaries.
func GetAppDir() (string, error) {
	var baseDir string
	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, "AppData", "Local")
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, ".local", "share")
	}

	appDir := filepath.Join(baseDir, "hittakeMD")
	// Check if old ytdl dir exists and hittakeMD does not, migrate bin
	oldAppDir := filepath.Join(baseDir, "ytdl")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if _, err := os.Stat(oldAppDir); err == nil {
			_ = os.Rename(oldAppDir, appDir)
		}
	}

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return appDir, nil
}

// GetBinDir returns the directory where portable binaries (yt-dlp, ffmpeg) are stored.
func GetBinDir() (string, error) {
	appDir, err := GetAppDir()
	if err != nil {
		return "", err
	}
	binDir := filepath.Join(appDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", err
	}
	return binDir, nil
}

// GetDefaultDownloadDir returns the default folder for storing downloaded videos.
func GetDefaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "downloads"
	}
	dlDir := filepath.Join(home, "Downloads", "hittakeMD")
	_ = os.MkdirAll(dlDir, 0755)
	return dlDir
}

// OpenFolder opens the given folder or parent directory of a file in the system file explorer.
func OpenFolder(targetPath string) error {
	if targetPath == "" {
		targetPath = GetDefaultDownloadDir()
	}
	targetPath = filepath.Clean(targetPath)

	folder := targetPath
	fi, err := os.Stat(targetPath)
	if err == nil && !fi.IsDir() {
		folder = filepath.Dir(targetPath)
	} else if err != nil {
		dir := filepath.Dir(targetPath)
		if fiDir, errDir := os.Stat(dir); errDir == nil && fiDir.IsDir() {
			folder = dir
		} else {
			folder = GetDefaultDownloadDir()
		}
	}

	folder = filepath.Clean(folder)
	_ = os.MkdirAll(folder, 0755)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		winFolder := filepath.FromSlash(folder)
		cmd = exec.Command("explorer.exe", winFolder)
	case "darwin":
		cmd = exec.Command("open", folder)
	default:
		cmd = exec.Command("xdg-open", folder)
	}
	return cmd.Start()
}

// FormatBytes formats byte counts into human readable strings (KiB, MiB, GiB).
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
