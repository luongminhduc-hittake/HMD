package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/downloader"
)

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
	presets := m.AvailablePresets
	if len(presets) == 0 {
		presets = downloader.AvailablePresets
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.PresetIndex > 0 {
				m.PresetIndex--
			}
		case "down", "j":
			if m.PresetIndex < len(presets)-1 {
				m.PresetIndex++
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(presets) {
				m.PresetIndex = idx
			}
		case "s", "S":
			if m.MediaInfo != nil && m.MediaInfo.IsImage {
				return m, nil
			}
			m.EnableSubtitles = !m.EnableSubtitles
			return m, nil
		case "t", "T":
			if m.MediaInfo != nil && m.MediaInfo.IsImage {
				return m, nil
			}
			m.State = StateInputTrim
			m.TrimInput.SetValue(m.TrimRange)
			m.TrimInput.Focus()
			return m, textinput.Blink
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

func (m Model) updateInputTrim(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.TrimRange = strings.TrimSpace(m.TrimInput.Value())
			m.State = StateSelectPreset
			return m, nil
		case "esc":
			m.State = StateSelectPreset
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.TrimInput, cmd = m.TrimInput.Update(msg)
	return m, cmd
}

func (m Model) viewSelectPlaylist() string {
	var s strings.Builder
	if m.MediaInfo != nil && m.MediaInfo.IsSpotify {
		s.WriteString(StyleBadgeSuccess.Render("SPOTIFY") + " Phát hiện Album / Playlist Spotify!\n\n")
		s.WriteString(fmt.Sprintf("Tiêu đề: %s\n", StyleHighlight.Render(m.MediaInfo.Title)))
		s.WriteString(StyleHelp.Render(fmt.Sprintf("Nghệ sĩ: %s • Số lượng: %d bài hát", m.MediaInfo.Uploader, len(m.MediaInfo.SpotifyTracks))) + "\n\n")
	} else {
		s.WriteString(StyleBadgeInfo.Render("PLAYLIST") + " Phát hiện Danh sách phát trong đường dẫn!\n\n")
		if m.MediaInfo != nil {
			s.WriteString(fmt.Sprintf("Tiêu đề: %s\n\n", StyleHighlight.Render(m.MediaInfo.Title)))
		}
	}

	var options []string
	if m.MediaInfo != nil && m.MediaInfo.IsSpotify {
		options = []string{
			"🎵  Chỉ tải bài đầu tiên",
			fmt.Sprintf("📑  Tải toàn bộ Album / Danh sách (%d bài hát)", len(m.MediaInfo.SpotifyTracks)),
		}
	} else {
		options = []string{
			"🎬  Chỉ tải video hiện tại",
			"📑  Tải toàn bộ danh sách phát (Playlist)",
		}
	}

	for i, opt := range options {
		if i == m.PlaylistIndex {
			s.WriteString(StyleSelected.Render(fmt.Sprintf("▶ [%d] %s", i+1, opt)) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(fmt.Sprintf("  [%d] %s", i+1, opt)) + "\n")
		}
	}

	s.WriteString("\n" + StyleHelp.Render("Phím tắt: ↑/↓ hoặc 1/2 để chọn • Enter: Xác nhận • Esc: Quay lại"))
	return StyleCard.Render(s.String())
}

func (m Model) viewSelectPreset() string {
	var s strings.Builder
	isImage := m.MediaInfo != nil && (m.MediaInfo.IsImage || m.MediaInfo.MediaType == downloader.MediaTypeImage)
	if m.MediaInfo != nil {
		icon := "🎬"
		role := "Kênh"
		if m.MediaInfo.IsSpotify || m.MediaInfo.MediaType == downloader.MediaTypeAudio {
			icon = "🎵"
			role = "Nghệ sĩ"
		} else if isImage {
			icon = "🖼️"
			role = "Tác giả"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", icon, StyleHighlight.Render(m.MediaInfo.Title)))
		if isImage {
			s.WriteString(StyleHelp.Render(fmt.Sprintf("👤 %s: %s", role, m.MediaInfo.Uploader)) + "\n\n")
		} else {
			s.WriteString(StyleHelp.Render(fmt.Sprintf("👤 %s: %s  •  ⏱ Thời lượng: %s", role, m.MediaInfo.Uploader, m.MediaInfo.Duration)) + "\n\n")
		}
	}

	if !isImage {
		trimStatus := "Không (Tải trọn vẹn)"
		if m.TrimRange != "" {
			trimStatus = m.TrimRange
		}
		subStatus := "TẮT"
		if m.EnableSubtitles {
			subStatus = "BẬT (vi, en soft-subs)"
		}
		s.WriteString(fmt.Sprintf("✂️  Cắt đoạn [t]: %s   │   💬 Phụ đề [s]: %s\n\n",
			StyleHighlight.Render(trimStatus), StyleHighlight.Render(subStatus)))
	}

	presets := m.AvailablePresets
	if len(presets) == 0 {
		presets = downloader.AvailablePresets
	}

	s.WriteString("Chọn định dạng và chất lượng tải về:\n\n")

	for i, opt := range presets {
		prefix := "  "
		line := fmt.Sprintf("[%d] %-22s │ %s", i+1, opt.Title, opt.Description)
		if i == m.PresetIndex {
			prefix = "▶ "
			s.WriteString(StyleSelected.Render(prefix+line) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(prefix+line) + "\n")
		}
	}

	shortcuts := fmt.Sprintf("Phím tắt: ↑/↓/1-%d: Chọn • Enter: Tải ngay • Esc: Quay lại", len(presets))
	if !isImage {
		shortcuts = fmt.Sprintf("Phím tắt: ↑/↓/1-%d: Chọn • t: Cắt đoạn • s: Bật/tắt phụ đề • Enter: Tải ngay • Esc: Quay lại", len(presets))
	}

	s.WriteString("\n" + StyleHelp.Render(shortcuts))
	return StyleCard.Render(s.String())
}

func (m Model) viewInputTrim() string {
	var s strings.Builder

	s.WriteString(StyleBadgeInfo.Render("CẮT ĐOẠN") + " " + StyleHighlight.Render("Cắt đoạn thời gian Video / Audio") + "\n\n")
	s.WriteString(StyleHelp.Render("Nhập khoảng thời gian cần cắt (VD: 01:20-03:45, 00:30-01:00 hoặc 0-60):\n(Để trống và nhấn Enter để hủy cắt đoạn, tải toàn bộ)\n\n"))
	s.WriteString(m.TrimInput.View() + "\n\n")
	s.WriteString(StyleHelp.Render("Phím tắt: Enter (Lưu & Quay lại) • Esc (Hủy bỏ)"))

	return StyleCard.Render(s.String())
}
