package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"hmd/internal/deps"
	"hmd/internal/downloader"
	"hmd/internal/updater"
	"hmd/internal/util"
)

// SessionState tracks the current view.
type SessionState int

const (
	StateCheckDeps SessionState = iota
	StateInputURL
	StateChangeDir
	StateFetchingInfo
	StateSelectPlaylist
	StateSelectPreset
	StateDownloading
	StateCompleted
	StateError
	StateUpdating
	StateHistory
	StateInputTrim
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
	Input    textinput.Model
	DirInput textinput.Model

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

	// History state
	History          []util.HistoryEntry
	HistoryIndex     int
	DuplicateHistory *util.HistoryEntry

	// Trim & Subtitle options
	TrimRange       string
	TrimInput       textinput.Model
	EnableSubtitles bool

	// Cookies & Concurrency options
	CookiesBrowser      string
	ConcurrentFragments int
	MaxRetries          int
	FragmentRetries     int
}

// InitialModel constructs the default model state.
func InitialModel(customOutputDir, initialURL string, customCookies ...string) Model {
	cfg := util.LoadConfig()
	effectiveDir := customOutputDir
	if effectiveDir == "" {
		effectiveDir = cfg.DownloadDir
	}

	cookies := cfg.CookiesBrowser
	if len(customCookies) > 0 && customCookies[0] != "" {
		cookies = customCookies[0]
	}

	hist := util.LoadHistory()
	ti := textinput.New()
	ti.Placeholder = "Dán đường dẫn YouTube tại đây (Ctrl+V hoặc chuột phải)..."
	var dup *util.HistoryEntry
	if initialURL != "" {
		ti.SetValue(initialURL)
		dup = util.FindHistoryInEntries(hist, initialURL)
	}
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60

	di := textinput.New()
	di.Placeholder = "Nhập đường dẫn thư mục lưu tải về..."
	di.SetValue(effectiveDir)
	di.CharLimit = 512
	di.Width = 60

	tri := textinput.New()
	tri.Placeholder = "VD: 01:20-03:45 hoặc 00:30-01:00 (Enter để lưu, để trống để tắt)..."
	tri.CharLimit = 64
	tri.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleHighlight

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(48),
	)

	return Model{
		State:               StateCheckDeps,
		OutputDir:           effectiveDir,
		Input:               ti,
		DirInput:            di,
		TrimInput:           tri,
		Spinner:             s,
		ProgressModel:       prog,
		PresetIndex:         0,
		PlaylistIndex:       0,
		ActionIndex:         0,
		History:             hist,
		DuplicateHistory:    dup,
		CookiesBrowser:      cookies,
		ConcurrentFragments: cfg.ConcurrentFragments,
		MaxRetries:          cfg.MaxRetries,
		FragmentRetries:     cfg.FragmentRetries,
	}
}
