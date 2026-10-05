//go:build windows

package downloader

import (
	"fmt"
	"os/exec"
)

// prepareCommand configures process termination on Windows using taskkill /T
// to ensure child processes (like ffmpeg.exe) are killed when the context is cancelled.
func prepareCommand(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		if cmd.Process == nil || cmd.Process.Pid <= 0 {
			return nil
		}
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(cmd.Process.Pid))
		if err := killCmd.Run(); err != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
}
