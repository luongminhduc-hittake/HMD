//go:build !windows

package downloader

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

// prepareCommand configures process group attributes to ensure child processes (like ffmpeg)
// are killed cleanly when the context is cancelled.
func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Give child process 2 seconds to terminate cleanly before SIGKILL
	cmd.WaitDelay = 2 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil || cmd.Process.Pid <= 0 {
			return nil
		}
		// Send SIGINT first to allow yt-dlp to clean up partial files
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		if err != nil && !errors.Is(err, syscall.ESRCH) {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
}
