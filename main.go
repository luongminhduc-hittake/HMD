package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"hmd/internal/tui"
	"hmd/internal/updater"
	"hmd/internal/util"
)

func main() {
	// 1. Startup cleanup: remove leftover .old executable from previous updates
	if execPath, err := os.Executable(); err == nil {
		_ = os.Remove(execPath + ".old")
	}

	// 2. Flags handling
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			fmt.Printf("hittakeMD v%s\n", updater.CurrentVersion)
			return
		}
		if arg == "--install" {
			if runtime.GOOS == "windows" {
				target, err := util.InstallOnWindows()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Lỗi cài đặt: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Đã cài đặt hittakeMD vào: %s\n", target)
				return
			}
			fmt.Println("--install chỉ áp dụng trên Windows (trên Linux hãy dùng ./install.sh)")
			return
		}
	}

	// 3. Initial URL from arguments or auto-detect from clipboard
	initialURL := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		initialURL = os.Args[1]
	} else {
		initialURL = util.GetMediaURLFromClipboard()
	}

	// Interactive TUI mode
	p := tea.NewProgram(
		tui.InitialModel("", initialURL),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi thực thi giao diện: %v\n", err)
		os.Exit(1)
	}
}
