package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ytdownloader/internal/tui"
	"ytdownloader/internal/util"
)

func main() {
	initialURL := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		initialURL = os.Args[1]
	}

	// Interactive TUI mode
	p := tea.NewProgram(
		tui.InitialModel(util.GetDefaultDownloadDir(), initialURL),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi thực thi giao diện: %v\n", err)
		os.Exit(1)
	}
}
