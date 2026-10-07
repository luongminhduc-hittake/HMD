package downloader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// execCommandContext allows mocking CLI execution during tests.
var execCommandContext = exec.CommandContext

// streamCommand executes a CLI command with process group isolation,
// merging stdout and stderr, and streaming output line-by-line via onLine callback.
func streamCommand(ctx context.Context, binPath string, args []string, onLine func(line string)) error {
	cmd := execCommandContext(ctx, binPath, args...)
	prepareCommand(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("không thể tạo pipe stdout: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("không thể khởi chạy tiến trình: %w", err)
	}

	reader := bufio.NewReaderSize(stdout, 64*1024)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "" && onLine != nil {
			onLine(line)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			break
		}
	}

	return cmd.Wait()
}
