package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Nord & Cyber Palette
	ColorNordPolarDark   = lipgloss.Color("#2E3440")
	ColorNordSnow        = lipgloss.Color("#ECEFF4")
	ColorNordSnowMuted   = lipgloss.Color("#D8DEE9")
	ColorNordFrostCyan   = lipgloss.Color("#88C0D0")
	ColorNordFrostBlue   = lipgloss.Color("#81A1C1")
	ColorNordFrostDeep   = lipgloss.Color("#5E81AC")
	ColorNordAuroraGreen = lipgloss.Color("#A3BE8C")
	ColorNordAuroraRed   = lipgloss.Color("#BF616A")
	ColorNordAuroraGold  = lipgloss.Color("#EBCB8B")
	ColorMuted           = lipgloss.Color("#616E88")

	// Component Styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorNordFrostCyan)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorNordFrostBlue)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorNordFrostDeep).
			Padding(1, 2)

	StyleSuccessCard = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorNordAuroraGreen).
				Padding(1, 2)

	StyleErrorCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorNordAuroraRed).
			Padding(1, 2)

	StyleBadgeSuccess = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorNordPolarDark).
				Background(ColorNordAuroraGreen).
				Padding(0, 1)

	StyleBadgeError = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorNordPolarDark).
			Background(ColorNordAuroraRed).
			Padding(0, 1)

	StyleBadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorNordPolarDark).
			Background(ColorNordFrostCyan).
			Padding(0, 1)

	StyleBadgeWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorNordPolarDark).
			Background(ColorNordAuroraGold).
			Padding(0, 1)

	StyleSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorNordSnow).
			Background(ColorNordFrostDeep).
			Padding(0, 1)

	StyleNormal = lipgloss.NewStyle().
			Foreground(ColorNordSnowMuted).
			Padding(0, 1)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleHighlight = lipgloss.NewStyle().
			Foreground(ColorNordFrostCyan).
			Bold(true)
)
