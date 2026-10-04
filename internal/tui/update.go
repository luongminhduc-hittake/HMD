package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"ytdownloader/internal/deps"
	"ytdownloader/internal/downloader"
	"ytdownloader/internal/updater"
	"ytdownloader/internal/util"
)

// Internal message types
type msgDepsProgress struct {
	item       string
	downloaded int64
	total      int64
	percent    float64
}

type msgDepsReady struct {
	paths deps.BinaryPaths
	err   error
}

type msgInfoFetched struct {
	info *downloader.MediaInfo
	err  error
}

type msgDownloadProgress struct {
	progress downloader.ProgressUpdate
}

type msgDownloadStatus struct {
	status string
}

type msgDownloadCompleted struct {
	result *downloader.DownloadResult
	err    error
}

type msgUpdateCheckResult struct {
	info *updater.ReleaseInfo
	err  error
}

type msgUpdateProgress struct {
	downloaded int64
	total      int64
	percent    float64
}

type msgUpdateFinished struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		initDepsCmd(),
		checkUpdateCmd(),
	)
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := updater.CheckForUpdate()
		return msgUpdateCheckResult{info: info, err: err}
	}
}

func startSelfUpdateCmd(assetURL string) (tea.Cmd, chan tea.Msg) {
	ch := make(chan tea.Msg, 50)
	go func() {
		err := updater.ApplyUpdate(assetURL, func(dl, total int64, pct float64) {
			ch <- msgUpdateProgress{downloaded: dl, total: total, percent: pct}
		})
		ch <- msgUpdateFinished{err: err}
		close(ch)
	}()
	return waitForChannel(ch), ch
}

func initDepsCmd() tea.Cmd {
	return func() tea.Msg {
		paths, missing := deps.FindBinaries()
		if len(missing) == 0 {
			return msgDepsReady{paths: paths, err: nil}
		}

		// Download missing dependencies
		readyPaths, err := deps.EnsureDependencies(nil)
		return msgDepsReady{paths: readyPaths, err: err}
	}
}

func fetchInfoCmd(ytdlpPath, url string) tea.Cmd {
	return func() tea.Msg {
		info, err := downloader.FetchInfo(ytdlpPath, url)
		return msgInfoFetched{info: info, err: err}
	}
}

func startDownloadCmd(m *Model) (tea.Cmd, chan tea.Msg) {
	ch := make(chan tea.Msg, 50)
	ctx, cancel := context.WithCancel(context.Background())
	m.DownloadCancel = cancel

	preset := downloader.AvailablePresets[m.PresetIndex].ID
	rawURL := m.MediaInfo.RawURL
	outDir := m.OutputDir
	if outDir == "" {
		outDir = util.GetDefaultDownloadDir()
	}

	opts := downloader.DownloadOptions{
		YtDlpPath:        m.Paths.YtDlp,
		FFmpegPath:       m.Paths.FFmpeg,
		URL:              rawURL,
		OutputDir:        outDir,
		Preset:           preset,
		DownloadPlaylist: m.DownloadPlaylist,
		OnProgress: func(pu downloader.ProgressUpdate) {
			ch <- msgDownloadProgress{progress: pu}
		},
		OnStatus: func(st string) {
			ch <- msgDownloadStatus{status: st}
		},
	}

	go func() {
		res, err := downloader.ExecuteDownload(ctx, opts)
		ch <- msgDownloadCompleted{result: res, err: err}
		close(ch)
	}()

	return waitForChannel(ch), ch
}

func waitForChannel(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.State == StateDownloading && m.DownloadCancel != nil {
				m.DownloadCancel()
				m.State = StateError
				m.ErrorMessage = "Quá trình tải xuống đã bị hủy bởi người dùng."
				m.ActionIndex = 0
				return m, nil
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		m.ProgressModel.Width = msg.Width - 20
		if m.ProgressModel.Width > 64 {
			m.ProgressModel.Width = 64
		}
		if m.ProgressModel.Width < 24 {
			m.ProgressModel.Width = 24
		}

	case spinner.TickMsg:
		var sCmd tea.Cmd
		m.Spinner, sCmd = m.Spinner.Update(msg)
		cmds = append(cmds, sCmd)

	case msgDepsReady:
		if msg.err != nil {
			m.State = StateError
			m.ErrorMessage = msg.err.Error()
			m.ActionIndex = 0
			return m, nil
		}
		m.Paths = msg.paths
		m.State = StateInputURL
		m.Input.Focus()
		return m, textinput.Blink

	case msgInfoFetched:
		if msg.err != nil {
			m.State = StateError
			m.ErrorMessage = msg.err.Error()
			m.ActionIndex = 0
			return m, nil
		}
		m.MediaInfo = msg.info
		if m.MediaInfo.IsPlaylist {
			m.State = StateSelectPlaylist
			m.PlaylistIndex = 0
		} else {
			m.State = StateSelectPreset
			m.PresetIndex = 0
		}
		return m, nil

	case msgDownloadProgress:
		m.ProgressData = msg.progress
		cmds = append(cmds, waitForChannel(m.ProgressChan))

	case msgDownloadStatus:
		m.ProgressData.StatusMessage = msg.status
		cmds = append(cmds, waitForChannel(m.ProgressChan))

	case msgDownloadCompleted:
		if msg.err != nil {
			m.State = StateError
			m.ErrorMessage = msg.err.Error()
			m.ActionIndex = 0
			return m, nil
		}
		m.Result = msg.result
		if m.MediaInfo != nil {
			m.Result.Title = m.MediaInfo.Title
		}
		m.State = StateCompleted
		m.ActionIndex = 0
		return m, nil

	case msgUpdateCheckResult:
		if msg.err == nil && msg.info != nil {
			m.AvailableUpdate = msg.info
		}
		return m, nil

	case msgUpdateProgress:
		m.UpdateProgress = msg.percent
		if msg.total > 0 {
			m.UpdateStatus = fmt.Sprintf("Đang tải bản cập nhật: %.1f%% (%d/%d MB)...", msg.percent, msg.downloaded/(1024*1024), msg.total/(1024*1024))
		}
		cmds = append(cmds, waitForChannel(m.ProgressChan))

	case msgUpdateFinished:
		if msg.err != nil {
			m.UpdateError = msg.err
		} else {
			m.UpdateSuccess = true
		}
		return m, nil
	}

	// State-specific interactions
	switch m.State {
	case StateInputURL:
		m, cmd = m.updateInputURL(msg)
		cmds = append(cmds, cmd)

	case StateSelectPlaylist:
		m, cmd = m.updateSelectPlaylist(msg)
		cmds = append(cmds, cmd)

	case StateSelectPreset:
		m, cmd = m.updateSelectPreset(msg)
		cmds = append(cmds, cmd)

	case StateCompleted:
		m, cmd = m.updateCompleted(msg)
		cmds = append(cmds, cmd)

	case StateError:
		m, cmd = m.updateError(msg)
		cmds = append(cmds, cmd)

	case StateUpdating:
		m, cmd = m.updateSelfUpdating(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateInputURL(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "u", "U":
			if m.AvailableUpdate != nil {
				m.State = StateUpdating
				m.UpdateStatus = fmt.Sprintf("Đang chuẩn bị tải bản cập nhật %s...", m.AvailableUpdate.TagName)
				m.UpdateProgress = 0
				m.UpdateError = nil
				m.UpdateSuccess = false
				var upCmd tea.Cmd
				upCmd, m.ProgressChan = startSelfUpdateCmd(m.AvailableUpdate.DownloadURL)
				return m, upCmd
			}
		case "enter":
			val := m.Input.Value()
			if val == "" {
				return m, nil
			}
			m.State = StateFetchingInfo
			return m, fetchInfoCmd(m.Paths.YtDlp, val)
		case "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m Model) updateSelectPlaylist(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.PlaylistIndex > 0 {
				m.PlaylistIndex--
			}
		case "down", "j":
			if m.PlaylistIndex < 1 {
				m.PlaylistIndex++
			}
		case "1":
			m.PlaylistIndex = 0
		case "2":
			m.PlaylistIndex = 1
		case "enter":
			m.DownloadPlaylist = (m.PlaylistIndex == 1)
			m.State = StateSelectPreset
			m.PresetIndex = 0
		case "esc":
			m.State = StateInputURL
			m.Input.Focus()
		}
	}
	return m, nil
}

func (m Model) updateSelectPreset(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.PresetIndex > 0 {
				m.PresetIndex--
			}
		case "down", "j":
			if m.PresetIndex < len(downloader.AvailablePresets)-1 {
				m.PresetIndex++
			}
		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(downloader.AvailablePresets) {
				m.PresetIndex = idx
			}
		case "enter":
			m.State = StateDownloading
			m.ProgressData = downloader.ProgressUpdate{
				StatusMessage: "Đang khởi tạo kết nối...",
			}
			var dlCmd tea.Cmd
			dlCmd, m.ProgressChan = startDownloadCmd(&m)
			return m, dlCmd
		case "esc":
			if m.MediaInfo != nil && m.MediaInfo.IsPlaylist {
				m.State = StateSelectPlaylist
			} else {
				m.State = StateInputURL
				m.Input.Focus()
			}
		}
	}
	return m, nil
}

func (m Model) updateCompleted(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.ActionIndex > 0 {
				m.ActionIndex--
			}
		case "down", "j":
			if m.ActionIndex < 2 {
				m.ActionIndex++
			}
		case "1":
			m.ActionIndex = 0
		case "2":
			m.ActionIndex = 1
		case "3":
			m.ActionIndex = 2
		case "enter":
			switch m.ActionIndex {
			case 0: // Tải tiếp video khác
				m.State = StateInputURL
				m.Input.SetValue("")
				m.Input.Focus()
				m.MediaInfo = nil
				m.Result = nil
				m.ProgressData = downloader.ProgressUpdate{}
				return m, nil
			case 1: // Mở thư mục chứa file
				target := ""
				if m.Result != nil && m.Result.FilePath != "" {
					target = m.Result.FilePath
				} else if m.OutputDir != "" {
					target = m.OutputDir
				} else {
					target = util.GetDefaultDownloadDir()
				}
				_ = util.OpenFolder(target)
				return m, nil
			case 2: // Thoát
				return m, tea.Quit
			}
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateError(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.ActionIndex > 0 {
				m.ActionIndex--
			}
		case "down", "j":
			if m.ActionIndex < 1 {
				m.ActionIndex++
			}
		case "1":
			m.ActionIndex = 0
		case "2":
			m.ActionIndex = 1
		case "enter":
			switch m.ActionIndex {
			case 0: // Thử lại
				m.State = StateInputURL
				m.Input.Focus()
				return m, nil
			case 1: // Thoát
				return m, tea.Quit
			}
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateSelfUpdating(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q":
			if m.UpdateSuccess {
				return m, tea.Quit
			}
			if m.UpdateError != nil {
				m.State = StateInputURL
				m.Input.Focus()
				return m, nil
			}
		}
	}
	return m, nil
}
