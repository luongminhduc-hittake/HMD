//go:build !windows

package util

// GetWindowsInstallDir returns empty string on non-Windows platforms.
func GetWindowsInstallDir() string {
	return ""
}

// IsInstalledOnWindows returns true on non-Windows platforms.
func IsInstalledOnWindows() bool {
	return true
}

// InstallOnWindows is a no-op on non-Windows platforms.
func InstallOnWindows() (string, error) {
	return "", nil
}
