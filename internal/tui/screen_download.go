package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewFetchingInfo() string {
	var s strings.Builder
	s.WriteString(m.Spinner.View() + " " + StyleHighlight.Render("Đang phân tích liên kết media...") + "\n\n")
	s.WriteString(StyleHelp.Render("Đang lấy thông tin định dạng và danh sách phát..."))
	return StyleCard.Render(s.String())
}

func (m Model) viewDownloading() string {
	var s strings.Builder
	if m.MediaInfo != nil {
		s.WriteString(fmt.Sprintf("🎬 %s\n\n", StyleHighlight.Render(m.MediaInfo.Title)))
	}

	// Progress bar
	pct := m.ProgressData.Percent
	if pct > 100 {
		pct = 100
	}
	s.WriteString(m.ProgressModel.ViewAs(pct / 100.0))
	s.WriteString(fmt.Sprintf("  %s\n\n", StyleHighlight.Render(fmt.Sprintf("%.1f%%", pct))))

	// Stats table
	speed := m.ProgressData.Speed
	if speed == "" {
		speed = "---"
	}
	eta := m.ProgressData.ETA
	if eta == "" {
		eta = "---"
	}
	total := m.ProgressData.TotalSize
	if total == "" {
		total = "---"
	}

	stats := fmt.Sprintf("⚡ Tốc độ: %-12s ⏳ Còn lại: %-10s 📦 Dung lượng: %s",
		StyleHighlight.Render(speed),
		StyleHighlight.Render(eta),
		StyleHighlight.Render(total),
	)
	s.WriteString(stats + "\n\n")

	// Current stage status
	statusMsg := m.ProgressData.StatusMessage
	if statusMsg == "" {
		statusMsg = "Đang kết nối đến máy chủ..."
	}
	s.WriteString(m.Spinner.View() + " " + StyleSubtitle.Render(statusMsg) + "\n\n")
	s.WriteString(StyleHelp.Render("Phím tắt: Ctrl+C để hủy tải"))

	return StyleCard.Render(s.String())
}
