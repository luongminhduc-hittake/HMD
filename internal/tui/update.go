package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/downloader"
	"hmd/internal/util"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		initDepsCmd(),
		checkUpdateCmd(),
	)
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
		if len(m.MediaInfo.AvailablePresets) > 0 {
			m.AvailablePresets = m.MediaInfo.AvailablePresets
		} else {
			m.AvailablePresets = downloader.AvailablePresets
		}
		m.PresetIndex = 0

		if m.MediaInfo.IsPlaylist {
			m.State = StateSelectPlaylist
			if m.MediaInfo.IsSpotify {
				m.PlaylistIndex = 1
			} else {
				m.PlaylistIndex = 0
			}
		} else {
			m.State = StateSelectPreset
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
		if m.Result != nil && m.MediaInfo != nil {
			presetTitle := "Mặc định"
			if len(m.AvailablePresets) > m.PresetIndex {
				presetTitle = m.AvailablePresets[m.PresetIndex].Title
			} else if len(downloader.AvailablePresets) > m.PresetIndex {
				presetTitle = downloader.AvailablePresets[m.PresetIndex].Title
			}
			_ = util.AddHistoryEntry(util.HistoryEntry{
				URL:       m.MediaInfo.RawURL,
				Title:     m.Result.Title,
				Format:    presetTitle,
				FilePath:  m.Result.FilePath,
				FileSize:  m.Result.FileSize,
				CreatedAt: time.Now(),
			})
			m.History = util.LoadHistory()
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

	case msgYtDlpUpdateFinished:
		if msg.err != nil {
			m.UpdateError = msg.err
		} else {
			m.Paths.YtDlp = msg.path
			m.UpdateSuccess = true
			m.UpdateStatus = "Đã cập nhật yt-dlp thành công!"
		}
		return m, nil
	}

	// State-specific interactions delegated to screen modules
	switch m.State {
	case StateInputURL:
		m, cmd = m.updateInputURL(msg)
		cmds = append(cmds, cmd)

	case StateChangeDir:
		m, cmd = m.updateChangeDir(msg)
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

	case StateHistory:
		m, cmd = m.updateHistory(msg)
		cmds = append(cmds, cmd)

	case StateInputTrim:
		m, cmd = m.updateInputTrim(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
