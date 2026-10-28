//go:build !windows

package util

// GetWindowsInstallDir returns empty string on non-Windows platforms.
func GetWindowsInstallDir() string {
	return ""
}

// InstallOnWindows is a no-op on non-Windows platforms.
func InstallOnWindows() (string, error) {
	return "", nil
}
