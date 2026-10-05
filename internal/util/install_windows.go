package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetWindowsInstallDir returns %LOCALAPPDATA%\Programs\hittakeMD
func GetWindowsInstallDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(localAppData, "Programs", "hittakeMD")
}

// IsInstalledOnWindows checks if the currently running executable is already in the install dir.
func IsInstalledOnWindows() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	execPath, err := os.Executable()
	if err != nil {
		return true
	}
	installDir := GetWindowsInstallDir()
	if installDir == "" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(installDir))
}

// InstallOnWindows copies the running executable to %LOCALAPPDATA%\Programs\hittakeMD\hmd.exe
// and creates Desktop & Start Menu shortcuts with the embedded icon.
func InstallOnWindows() (string, error) {
	if runtime.GOOS != "windows" {
		return "", nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("không thể xác định vị trí file chạy: %w", err)
	}

	installDir := GetWindowsInstallDir()
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("không thể tạo thư mục cài đặt: %w", err)
	}

	targetExe := filepath.Join(installDir, "hmd.exe")

	// If source is already target, skip copying
	if strings.ToLower(execPath) != strings.ToLower(targetExe) {
		src, err := os.Open(execPath)
		if err != nil {
			return "", fmt.Errorf("không thể đọc file nguồn: %w", err)
		}
		defer src.Close()

		dst, err := os.OpenFile(targetExe, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", fmt.Errorf("không thể tạo file đích: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return "", fmt.Errorf("lỗi sao chép file: %w", err)
		}
	}

	// Create Desktop and Start Menu shortcuts via PowerShell
	createShortcutsScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell

$Target = "%s"
$WorkingDir = "%s"

# Desktop Shortcut
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$ShortcutDesktop = $WshShell.CreateShortcut("$DesktopPath\hittakeMD.lnk")
$ShortcutDesktop.TargetPath = $Target
$ShortcutDesktop.WorkingDirectory = $WorkingDir
$ShortcutDesktop.Description = "hittake's media downloader"
$ShortcutDesktop.Save()

# Start Menu Shortcut
$StartMenuPath = [Environment]::GetFolderPath("Programs")
$ShortcutStart = $WshShell.CreateShortcut("$StartMenuPath\hittakeMD.lnk")
$ShortcutStart.TargetPath = $Target
$ShortcutStart.WorkingDirectory = $WorkingDir
$ShortcutStart.Description = "hittake's media downloader"
$ShortcutStart.Save()
`, strings.ReplaceAll(targetExe, `\`, `\\`), strings.ReplaceAll(installDir, `\`, `\\`))

	psCmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", createShortcutsScript)
	_ = psCmd.Run()

	return targetExe, nil
}
