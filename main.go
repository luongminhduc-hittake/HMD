package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ytdownloader/internal/cli"
	"ytdownloader/internal/tui"
	"ytdownloader/internal/util"
)

func main() {
	// If CLI arguments or flags are supplied, execute headless CLI mode
	if len(os.Args) > 1 {
		os.Exit(cli.Run(os.Args[1:]))
	}

	// Interactive TUI mode (Double-click or running `ytdl` without arguments)
	p := tea.NewProgram(
		tui.InitialModel(util.GetDefaultDownloadDir()),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi thực thi giao diện: %v\n", err)
		os.Exit(1)
	}
}
