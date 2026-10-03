package tui

import (
	"fmt"
	"strings"

	"ytdownloader/internal/downloader"
)

const BannerASCII = `
  ██╗   ██╗████████╗██████╗ ██╗
  ╚██╗ ██╔╝╚══██╔══╝██╔══██╗██║
   ╚████╔╝    ██║   ██║  ██║██║
    ╚██╔╝     ██║   ██║  ██║██║
     ██║      ██║   ██████╔╝███████╗
     ╚═╝      ╚═╝   ╚═════╝ ╚══════╝`

func (m Model) View() string {
	var b strings.Builder

	// Top Banner
	b.WriteString(StyleTitle.Render(BannerASCII))
	b.WriteString("\n")
	b.WriteString(StyleSubtitle.Render("    ⚡ Modern YouTube Downloader for Windows"))
	b.WriteString("\n\n")

	switch m.State {
	case StateCheckDeps:
		b.WriteString(m.viewCheckDeps())
	case StateInputURL:
		b.WriteString(m.viewInputURL())
	case StateFetchingInfo:
		b.WriteString(m.viewFetchingInfo())
	case StateSelectPlaylist:
		b.WriteString(m.viewSelectPlaylist())
	case StateSelectPreset:
		b.WriteString(m.viewSelectPreset())
	case StateDownloading:
		b.WriteString(m.viewDownloading())
	case StateCompleted:
		b.WriteString(m.viewCompleted())
	case StateError:
		b.WriteString(m.viewError())
	}

	return b.String()
}

func (m Model) viewCheckDeps() string {
	var s strings.Builder
	s.WriteString(StyleBadgeInfo.Render("KHỞI TẠO"))
	s.WriteString(" Đang kiểm tra công cụ cần thiết (yt-dlp, ffmpeg)...\n\n")

	if len(m.DepsMissing) > 0 {
		s.WriteString(fmt.Sprintf("Đang tải %s tự động (chỉ tải 1 lần duy nhất)...\n\n", StyleHighlight.Render(m.DepsItemName)))
		s.WriteString(m.ProgressModel.ViewAs(m.DepsPercent / 100.0))
		s.WriteString(fmt.Sprintf("  %.1f%%\n\n", m.DepsPercent))
	} else {
		s.WriteString(m.Spinner.View() + " Đang kiểm tra hệ thống...")
	}

	return StyleCard.Render(s.String())
}

func (m Model) viewInputURL() string {
	var s strings.Builder
	s.WriteString(StyleHighlight.Render("Nhập liên kết video hoặc danh sách phát:") + "\n\n")
	s.WriteString(m.Input.View() + "\n\n")

	if m.OutputDir != "" {
		s.WriteString(StyleHelp.Render(fmt.Sprintf("📂 Thư mục lưu: %s\n", m.OutputDir)))
	}
	s.WriteString(StyleHelp.Render("Phím tắt: Enter (Tiếp tục) • Esc / Ctrl+C (Thoát)"))

	return StyleCard.Render(s.String())
}

func (m Model) viewFetchingInfo() string {
	var s strings.Builder
	s.WriteString(m.Spinner.View() + " " + StyleHighlight.Render("Đang phân tích link YouTube...") + "\n\n")
	s.WriteString(StyleHelp.Render("Đang lấy thông tin định dạng và danh sách phát..."))
	return StyleCard.Render(s.String())
}

func (m Model) viewSelectPlaylist() string {
	var s strings.Builder
	s.WriteString(StyleBadgeInfo.Render("PLAYLIST") + " Phát hiện Danh sách phát trong đường dẫn!\n\n")
	if m.MediaInfo != nil {
		s.WriteString(fmt.Sprintf("Tiêu đề: %s\n\n", StyleHighlight.Render(m.MediaInfo.Title)))
	}

	options := []string{
		"🎬  Chỉ tải video hiện tại",
		"📑  Tải toàn bộ danh sách phát (Playlist)",
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
	if m.MediaInfo != nil {
		s.WriteString(fmt.Sprintf("🎬 %s\n", StyleHighlight.Render(m.MediaInfo.Title)))
		s.WriteString(StyleHelp.Render(fmt.Sprintf("👤 Kênh: %s  •  ⏱ Thời lượng: %s", m.MediaInfo.Uploader, m.MediaInfo.Duration)) + "\n\n")
	}

	s.WriteString("Chọn định dạng và chất lượng tải về:\n\n")

	for i, opt := range downloader.AvailablePresets {
		prefix := "  "
		line := fmt.Sprintf("[%d] %-22s │ %s", i+1, opt.Title, opt.Description)
		if i == m.PresetIndex {
			prefix = "▶ "
			s.WriteString(StyleSelected.Render(prefix+line) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(prefix+line) + "\n")
		}
	}

	s.WriteString("\n" + StyleHelp.Render("Phím tắt: ↑/↓ hoặc phím 1-5: Chọn • Enter: Tải ngay • Esc: Quay lại"))
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
		statusMsg = "Đang kết nối đến YouTube..."
	}
	s.WriteString(m.Spinner.View() + " " + StyleSubtitle.Render(statusMsg) + "\n\n")
	s.WriteString(StyleHelp.Render("Phím tắt: Ctrl+C để hủy tải"))

	return StyleCard.Render(s.String())
}

func (m Model) viewCompleted() string {
	var s strings.Builder
	s.WriteString(StyleBadgeSuccess.Render("✔ HOÀN TẤT THÀNH CÔNG") + "\n\n")

	if m.Result != nil {
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
		"❌  Thoát chương trình",
	}

	for i, act := range actions {
		if i == m.ActionIndex {
			s.WriteString(StyleSelected.Render(fmt.Sprintf("▶ [%d] %s", i+1, act)) + "\n")
		} else {
			s.WriteString(StyleNormal.Render(fmt.Sprintf("  [%d] %s", i+1, act)) + "\n")
		}
	}

	s.WriteString("\n" + StyleHelp.Render("Phím tắt: ↑/↓ hoặc 1/2 để chọn • Enter: Thực hiện"))
	return StyleErrorCard.Render(s.String())
}
