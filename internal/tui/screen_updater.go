package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSelfUpdating(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q":
			if m.UpdateSuccess {
				if m.UpdateStatus == "Đã cập nhật yt-dlp thành công!" {
					m.State = StateInputURL
					m.Input.Focus()
					return m, nil
				}
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

func (m Model) viewUpdating() string {
	var s strings.Builder
	s.WriteString(StyleBadgeInfo.Render("CẬP NHẬT") + " " + StyleHighlight.Render("Tự động nâng cấp phiên bản mới") + "\n\n")

	if m.UpdateSuccess {
		s.WriteString(StyleBadgeSuccess.Render("✔ CẬP NHẬT THÀNH CÔNG!") + "\n\n")
		if m.UpdateStatus != "" {
			s.WriteString(m.UpdateStatus + "\n\n")
		} else {
			s.WriteString("Ứng dụng đã được nâng cấp lên phiên bản mới nhất thành công.\n\n")
		}
		if m.UpdateStatus == "Đã cập nhật yt-dlp thành công!" {
			s.WriteString(StyleHelp.Render("Nhấn Enter hoặc Esc để tiếp tục."))
		} else {
			s.WriteString(StyleHelp.Render("Nhấn Enter hoặc Esc để thoát. Vui lòng mở lại hmd để sử dụng."))
		}
		return StyleSuccessCard.Render(s.String())
	}

	if m.UpdateError != nil {
		s.WriteString(StyleBadgeError.Render("✖ CẬP NHẬT THẤT BẠI") + "\n\n")
		s.WriteString(m.UpdateError.Error() + "\n\n")
		s.WriteString(StyleHelp.Render("Nhấn Enter hoặc Esc để quay lại màn hình chính."))
		return StyleErrorCard.Render(s.String())
	}

	s.WriteString(m.Spinner.View() + " " + m.UpdateStatus + "\n\n")
	if m.UpdateProgress > 0 {
		s.WriteString(m.ProgressModel.ViewAs(m.UpdateProgress / 100.0) + "\n\n")
	}
	s.WriteString(StyleHelp.Render("Đang cài đặt trực tiếp vào hệ thống... Vui lòng chờ."))

	return StyleCard.Render(s.String())
}
