//go:build !windows

package downloader

import (
	"os/exec"
	"syscall"
)

// prepareCommand configures process group attributes to ensure child processes (like ffmpeg)
// are killed cleanly when the context is cancelled.
func prepareCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil || cmd.Process.Pid <= 0 {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
