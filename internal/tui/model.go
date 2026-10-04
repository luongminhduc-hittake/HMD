package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"ytdownloader/internal/deps"
	"ytdownloader/internal/downloader"
	"ytdownloader/internal/updater"
)

// SessionState tracks the current view.
type SessionState int

const (
	StateCheckDeps SessionState = iota
	StateInputURL
	StateFetchingInfo
	StateSelectPlaylist
	StateSelectPreset
	StateDownloading
	StateCompleted
	StateError
	StateUpdating
)

// Model represents the overall TUI application state.
type Model struct {
	State        SessionState
	Paths        deps.BinaryPaths
	OutputDir    string
	TerminalWidth int
	TerminalHeight int

	// Dependency check
	DepsMissing  []string
	DepsItemName string
	DepsPercent  float64

	// Input view
	Input textinput.Model

	// Spinner
	Spinner spinner.Model

	// Media info
	MediaInfo *downloader.MediaInfo

	// Playlist choice
	PlaylistIndex   int
	DownloadPlaylist bool

	// Preset choice
	PresetIndex int

	// Download state
	ProgressModel progress.Model
	ProgressData  downloader.ProgressUpdate
	DownloadCancel context.CancelFunc
	ProgressChan   chan tea.Msg

	// Completion / Error state
	Result       *downloader.DownloadResult
	ActionIndex  int // for completed/error screen actions
	ErrorMessage string

	// Self-update state
	AvailableUpdate *updater.ReleaseInfo
	UpdateStatus    string
	UpdateProgress  float64
	UpdateError     error
	UpdateSuccess   bool
}

// InitialModel constructs the default model state.
func InitialModel(customOutputDir, initialURL string) Model {
	ti := textinput.New()
	ti.Placeholder = "Dán đường dẫn YouTube tại đây (Ctrl+V hoặc chuột phải)..."
	if initialURL != "" {
		ti.SetValue(initialURL)
	}
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleHighlight

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(48),
	)

	return Model{
		State:         StateCheckDeps,
		OutputDir:     customOutputDir,
		Input:         ti,
		Spinner:       s,
		ProgressModel: prog,
		PresetIndex:   0,
		PlaylistIndex: 0,
		ActionIndex:   0,
	}
}
