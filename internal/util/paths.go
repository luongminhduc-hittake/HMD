package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	if fi, err := os.Stat(targetPath); err != nil || !fi.IsDir() {
		folder = filepath.Dir(targetPath)
		if fiDir, err := os.Stat(folder); err != nil || !fiDir.IsDir() {
			folder = GetDefaultDownloadDir()
		}
	}

	_ = os.MkdirAll(folder, 0755)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		winFolder := filepath.FromSlash(folder)
		cmd = exec.Command("explorer.exe", winFolder)
	case "darwin":
		cmd = exec.Command("open", "--", folder)
	default:
		cmd = exec.Command("xdg-open", "--", folder)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
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

// SanitizeFilename removes illegal filesystem characters across Windows and Unix platforms.
func SanitizeFilename(name string) string {
	invalid := []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*'}
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if r < 32 {
			continue
		}
		isInvalid := false
		for _, inv := range invalid {
			if r == inv {
				isInvalid = true
				break
			}
		}
		if isInvalid {
			result = append(result, '_')
		} else {
			result = append(result, r)
		}
	}
	s := strings.TrimSpace(string(result))
	s = strings.Trim(s, ".")
	if s == "" {
		return "untitled"
	}
	return s
}
