package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ytdownloader/internal/tui"
	"ytdownloader/internal/updater"
)

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			fmt.Printf("ytdl v%s\n", updater.CurrentVersion)
			return
		}
	}

	initialURL := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		initialURL = os.Args[1]
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
