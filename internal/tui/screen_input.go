package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hmd/internal/downloader"
	"hmd/internal/util"
)

func (m Model) triggerPickFolder() (Model, tea.Cmd) {
	m.State = StateChangeDir
	m.DirInput.SetValue(m.OutputDir)
	m.DirInput.Focus()
	return m, textinput.Blink
}

func (m Model) updateInputURL(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+u":
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
		case "ctrl+b":
			m.CookiesBrowser = util.CycleBrowser(m.CookiesBrowser)
			cfg := util.LoadConfig()
			cfg.CookiesBrowser = m.CookiesBrowser
			_ = util.SaveConfig(cfg)
			return m, nil
		case "ctrl+o":
			return m.triggerPickFolder()
		case "ctrl+h":
			m.History = util.LoadHistory()
			m.HistoryIndex = 0
			m.State = StateHistory
			return m, nil
		case "ctrl+y":
			m.State = StateUpdating
			m.UpdateStatus = "Đang tải bản cập nhật yt-dlp mới nhất..."
			m.UpdateProgress = 0
			m.UpdateError = nil
			m.UpdateSuccess = false
			var upCmd tea.Cmd
			upCmd, m.ProgressChan = updateYtDlpCmd()
			return m, upCmd
		case "ctrl+p":
			if m.DuplicateHistory != nil && m.DuplicateHistory.FilePath != "" {
				_ = util.OpenFolder(m.DuplicateHistory.FilePath)
				return m, nil
			}
		case "enter":
			val := strings.TrimSpace(m.Input.Value())
			if val == "" {
				return m, nil
			}
			if !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
				m.State = StateError
				m.ErrorMessage = "Đường dẫn không hợp lệ. Vui lòng nhập link bắt đầu bằng http:// hoặc https://"
				m.ActionIndex = 0
				return m, nil
			}
			m.State = StateFetchingInfo
			return m, fetchInfoCmd(m.Paths.YtDlp, val, m.CookiesBrowser)
		case "esc":
			if m.Input.Value() != "" {
				m.Input.SetValue("")
				m.DuplicateHistory = nil
				return m, nil
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	prevVal := m.Input.Value()
	m.Input, cmd = m.Input.Update(msg)
	if m.Input.Value() != prevVal {
		m.DuplicateHistory = util.FindHistoryInEntries(m.History, m.Input.Value())
	}
	return m, cmd
}

func (m Model) updateChangeDir(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.DirInput.Value())
			if val != "" {
				_ = os.MkdirAll(val, 0755)
				m.OutputDir = val
				_ = util.SaveConfig(util.Config{
					DownloadDir:         val,
					CookiesBrowser:      m.CookiesBrowser,
					ConcurrentFragments: m.ConcurrentFragments,
					MaxRetries:          m.MaxRetries,
					FragmentRetries:     m.FragmentRetries,
				})
			}
			m.State = StateInputURL
			m.Input.Focus()
			return m, textinput.Blink
		case "esc":
			m.DirInput.SetValue(m.OutputDir)
			m.State = StateInputURL
			m.Input.Focus()
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.DirInput, cmd = m.DirInput.Update(msg)
	return m, cmd
}

func (m Model) viewInputURL() string {
	var s strings.Builder

	if m.AvailableUpdate != nil {
		s.WriteString(StyleBadgeWarning.Render("BẢN MỚI") + " " +
			StyleHighlight.Render(fmt.Sprintf("Có bản cập nhật mới %s! Nhấn Ctrl+U để nâng cấp tự động.", m.AvailableUpdate.TagName)) + "\n\n")
	}

	if m.DuplicateHistory != nil {
		s.WriteString(StyleBadgeWarning.Render("ĐÃ TẢI TRƯỚC ĐÂY") + " " +
			StyleHighlight.Render(fmt.Sprintf("Đã tải: %s (%s)", m.DuplicateHistory.Title, m.DuplicateHistory.Format)) + "\n" +
			StyleHelp.Render(fmt.Sprintf("📂 Tệp: %s\n⏱ Ngày tải: %s (Nhấn Ctrl+P để mở tệp ngay)", m.DuplicateHistory.FilePath, m.DuplicateHistory.CreatedAt.Format("15:04 02/01/2006"))) + "\n\n")
	}

	s.WriteString(StyleHighlight.Render("Nhập liên kết video, âm thanh hoặc danh sách phát:") + "\n\n")
	s.WriteString(m.Input.View() + "\n\n")

	currVal := strings.TrimSpace(m.Input.Value())
	if currVal != "" {
		p := downloader.DetectPlatform(currVal)
		badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color(p.Color)).Bold(true).Padding(0, 1)
		if p.Name == "Media" {
			s.WriteString(badgeStyle.Render(p.Name) + " " + StyleHelp.Render("Liên kết ngoài danh mục — cứ thử tải xem được không nhé!") + "\n\n")
		} else {
			s.WriteString(badgeStyle.Render(p.Name) + " " + StyleHelp.Render(fmt.Sprintf("Phát hiện liên kết từ %s", p.Name)) + "\n\n")
		}
	} else {
		s.WriteString(StyleHelp.Render("🌐 Hỗ trợ: YouTube • TikTok • Facebook • Instagram • X • SoundCloud • Spotify • Reddit • Pinterest\n💡 Trang khác: Cứ thử dán link vào xem tải được không nhé!") + "\n\n")
	}

	if m.OutputDir != "" {
		s.WriteString(StyleHelp.Render(fmt.Sprintf("📂 Thư mục lưu: %s  (Ctrl+O để đổi)\n", m.OutputDir)))
	}

	cookiesStatus := "TẮT"
	if strings.EqualFold(m.CookiesBrowser, "auto") {
		cookiesStatus = "AUTO (Tự động)"
	} else if m.CookiesBrowser != "" {
		cookiesStatus = strings.ToUpper(m.CookiesBrowser)
	}
	s.WriteString(StyleHelp.Render(fmt.Sprintf("🍪 Cookies trình duyệt: %s  (Ctrl+B để đổi)\n\n", StyleHighlight.Render(cookiesStatus))))

	helpText := "Enter: Tiếp tục • Ctrl+B: Cookies • Ctrl+H: Lịch sử • Ctrl+O: Thư mục • Ctrl+Y: Cập nhật yt-dlp • Esc: Xóa / Thoát"
	if m.DuplicateHistory != nil {
		helpText = "Enter: Tiếp tục • Ctrl+P: Mở tệp cũ • Ctrl+B: Cookies • Ctrl+H: Lịch sử • Ctrl+O: Thư mục • Ctrl+Y: Cập nhật yt-dlp • Esc: Xóa / Thoát"
	}
	if m.AvailableUpdate != nil {
		helpText += " • Ctrl+U: Cập nhật app"
	}
	s.WriteString(StyleHelp.Render(helpText))

	return StyleCard.Render(s.String())
}

func (m Model) viewChangeDir() string {
	var s strings.Builder

	s.WriteString(StyleHighlight.Render("Thay đổi thư mục lưu tải về:") + "\n\n")
	s.WriteString(m.DirInput.View() + "\n\n")
	s.WriteString(StyleHelp.Render("Phím tắt: Enter (Lưu & Áp dụng) • Esc (Hủy & Quay lại)"))

	return StyleCard.Render(s.String())
}
