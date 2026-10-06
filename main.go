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

	// 2. Flags & argument handling
	cookiesArg := ""
	initialURL := ""
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
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
		if (arg == "-cookies" || arg == "--cookies") && i+1 < len(os.Args) {
			cookiesArg = os.Args[i+1]
			i++
			continue
		}
		if !strings.HasPrefix(arg, "-") && initialURL == "" {
			initialURL = arg
		}
	}

	// 3. Initial URL auto-detect from clipboard if not passed via args
	if initialURL == "" {
		initialURL = util.GetMediaURLFromClipboard()
	}

	// Interactive TUI mode
	p := tea.NewProgram(
		tui.InitialModel("", initialURL, cookiesArg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi thực thi giao diện: %v\n", err)
		os.Exit(1)
	}
}
