package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/downloader"
	"hmd/internal/util"
)

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
			case 0: // Thử lại với link khác
				m.State = StateInputURL
				m.Input.Focus()
				return m, nil
			case 1: // Cập nhật yt-dlp mới nhất & thử lại
				m.State = StateUpdating
				m.UpdateStatus = "Đang tải bản cập nhật yt-dlp mới nhất..."
				m.UpdateProgress = 0
				m.UpdateError = nil
				m.UpdateSuccess = false
				var upCmd tea.Cmd
				upCmd, m.ProgressChan = updateYtDlpCmd()
				return m, upCmd
			case 2: // Thoát chương trình
				return m, tea.Quit
			}
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateHistory(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.HistoryIndex > 0 {
				m.HistoryIndex--
			}
		case "down", "j":
			if m.HistoryIndex < len(m.History)-1 {
				m.HistoryIndex++
			}
		case "enter", "o":
			if len(m.History) > 0 && m.HistoryIndex < len(m.History) {
				_ = util.OpenFolder(m.History[m.HistoryIndex].FilePath)
			}
		case "d", "backspace":
			if len(m.History) > 0 && m.HistoryIndex < len(m.History) {
				_ = util.DeleteHistoryEntry(m.HistoryIndex)
				m.History = util.LoadHistory()
				if m.HistoryIndex >= len(m.History) && m.HistoryIndex > 0 {
					m.HistoryIndex = len(m.History) - 1
				}
			}
		case "esc", "q", "h":
			m.State = StateInputURL
			m.Input.Focus()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewCompleted() string {
	var s strings.Builder
	s.WriteString(StyleBadgeSuccess.Render("✔ HOÀN TẤT THÀNH CÔNG") + "\n\n")

	if m.Result != nil {
		if m.Result.IsPlaylist {
			s.WriteString(fmt.Sprintf("Loại tải:   %s\n", StyleHighlight.Render("Danh sách phát (Playlist)")))
			if m.Result.Title != "" {
				s.WriteString(fmt.Sprintf("Tiêu đề:    %s\n", StyleHighlight.Render(m.Result.Title)))
			}
			if m.Result.DownloadedCount > 0 {
				s.WriteString(fmt.Sprintf("Đã tải:     %s\n", StyleHighlight.Render(fmt.Sprintf("%d mục thành công", m.Result.DownloadedCount))))
			}
			if m.Result.SkippedCount > 0 {
				s.WriteString(fmt.Sprintf("Bỏ qua:     %s\n", StyleBadgeWarning.Render(fmt.Sprintf("%d mục (bị ẩn hoặc không khả dụng)", m.Result.SkippedCount))))
			}
			if len(m.Result.FailedTracks) > 0 {
				s.WriteString("\n" + StyleHelp.Render("Danh sách các mục bị bỏ qua:") + "\n")
				limit := 4
				if len(m.Result.FailedTracks) < limit {
					limit = len(m.Result.FailedTracks)
				}
				for i := 0; i < limit; i++ {
					s.WriteString(StyleHelp.Render(fmt.Sprintf(" • %s", m.Result.FailedTracks[i])) + "\n")
				}
				if len(m.Result.FailedTracks) > limit {
					s.WriteString(StyleHelp.Render(fmt.Sprintf(" • ...và %d mục khác", len(m.Result.FailedTracks)-limit)) + "\n")
				}
				s.WriteString("\n")
			}
			if m.Result.FilePath != "" {
				s.WriteString(fmt.Sprintf("Thư mục:    %s\n\n", StyleSubtitle.Render(m.Result.FilePath)))
			}
		} else {
			if m.Result.Title != "" {
				s.WriteString(fmt.Sprintf("Tiêu đề:    %s\n", StyleHighlight.Render(m.Result.Title)))
			}
			if m.Result.FileName != "" {
				s.WriteString(fmt.Sprintf("Tên tệp:    %s\n", StyleHighlight.Render(m.Result.FileName)))
			}
			if m.Result.FormattedSize != "" {
				s.WriteString(fmt.Sprintf("Kích thước: %s\n", StyleHighlight.Render(m.Result.FormattedSize)))
			}
			if m.Result.FilePath != "" {
				s.WriteString(fmt.Sprintf("Lưu tại:    %s\n\n", StyleSubtitle.Render(m.Result.FilePath)))
			}
		}
	}

	actions := []string{
		"🔁  Tải tiếp video khác",
		"📂  Mở thư mục chứa file",
		"❌  Thoát chương trình",
	}

	for i, act := range actions {
		if i == m.ActionIndex {
			s.WriteString(StyleSelected.Render(fmt.Sprintf("▶ [%d] %s", i+1, act)) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(fmt.Sprintf("  [%d] %s", i+1, act)) + "\n")
		}
	}

	s.WriteString("\n" + StyleHelp.Render("Phím tắt: ↑/↓ hoặc 1/2/3 để chọn • Enter: Thực hiện"))
	return StyleSuccessCard.Render(s.String())
}

func (m Model) viewError() string {
	var s strings.Builder
	s.WriteString(StyleBadgeError.Render("✖ ĐÃ XẢY RA LỖI") + "\n\n")
	s.WriteString(m.ErrorMessage + "\n\n")

	actions := []string{
		"🔁  Thử lại với link khác",
		"⚡  Cập nhật yt-dlp mới nhất & thử lại",
		"❌  Thoát chương trình",
	}

	for i, act := range actions {
		if i == m.ActionIndex {
			s.WriteString(StyleSelected.Render(fmt.Sprintf("▶ [%d] %s", i+1, act)) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(fmt.Sprintf("  [%d] %s", i+1, act)) + "\n")
		}
	}

	s.WriteString("\n" + StyleHelp.Render("Phím tắt: ↑/↓ hoặc 1/2/3 để chọn • Enter: Thực hiện"))
	return StyleErrorCard.Render(s.String())
}

func (m Model) viewHistory() string {
	var s strings.Builder

	s.WriteString(StyleBadgeInfo.Render("LỊCH SỬ TẢI") + " " + StyleHighlight.Render(fmt.Sprintf("Lịch sử tải xuống gần đây (%d tệp)", len(m.History))) + "\n\n")

	if len(m.History) == 0 {
		s.WriteString(StyleHelp.Render("Chưa có lượt tải nào được ghi nhận.") + "\n\n")
	} else {
		start := 0
		maxVisible := 7
		if m.HistoryIndex >= maxVisible {
			start = m.HistoryIndex - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(m.History) {
			end = len(m.History)
		}

		for i := start; i < end; i++ {
			entry := m.History[i]
			sizeStr := ""
			if entry.FileSize > 0 {
				sizeStr = fmt.Sprintf(" • %.1f MB", float64(entry.FileSize)/(1024*1024))
			}
			dateStr := entry.CreatedAt.Format("15:04 02/01/06")
			prefix := "  "
			title := entry.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			line := fmt.Sprintf("[%d] %s (%s%s - %s)", i+1, title, entry.Format, sizeStr, dateStr)
			if i == m.HistoryIndex {
				prefix = "▶ "
				s.WriteString(StyleSelected.Render(prefix+line) + "\n")
				s.WriteString(StyleHelp.Render(fmt.Sprintf("    📂 %s", entry.FilePath)) + "\n")
			} else {
				s.WriteString(StyleNormal.Render(prefix+line) + "\n")
			}
		}
		s.WriteString("\n")
	}

	s.WriteString(StyleHelp.Render("Phím tắt: ↑/↓ hoặc j/k (Chọn) • Enter/o (Mở tệp/thư mục) • d (Xóa dòng này) • Esc/h (Quay lại)"))
	return StyleCard.Render(s.String())
}
