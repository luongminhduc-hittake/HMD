package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/downloader"
	"hmd/internal/util"
)

func TestModelTrimAndSubtitles(t *testing.T) {
	m := InitialModel("", "")
	m.State = StateSelectPreset
	m.MediaInfo = &downloader.MediaInfo{
		Title:    "Test Video",
		Uploader: "Artist",
		Duration: "03:45",
	}

	// Toggle subtitle with "s"
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model2 := m2.(Model)
	if !model2.EnableSubtitles {
		t.Errorf("expected EnableSubtitles to be true after pressing 's'")
	}

	// Check view shows subtitle ON
	view2 := model2.View()
	if !strings.Contains(view2, "BẬT (vi, en soft-subs)") {
		t.Errorf("expected view to reflect subtitles enabled, got: %s", view2)
	}

	// Switch to trim input with "t"
	m3, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model3 := m3.(Model)
	if model3.State != StateInputTrim {
		t.Errorf("expected state StateInputTrim after pressing 't', got: %v", model3.State)
	}

	// Enter trim value and press enter
	model3.TrimInput.SetValue("01:20-03:45")
	m4, _ := model3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model4 := m4.(Model)
	if model4.State != StateSelectPreset {
		t.Errorf("expected return to StateSelectPreset, got: %v", model4.State)
	}
	if model4.TrimRange != "01:20-03:45" {
		t.Errorf("expected TrimRange '01:20-03:45', got %q", model4.TrimRange)
	}

	// Check view shows trim
	view4 := model4.View()
	if !strings.Contains(view4, "01:20-03:45") {
		t.Errorf("expected view to show trim range, got: %s", view4)
	}
}

func TestModelHistoryView(t *testing.T) {
	m := InitialModel("", "")
	m.State = StateHistory
	m.History = []util.HistoryEntry{
		{
			URL:       "https://youtube.com/watch?v=123",
			Title:     "Sample Song",
			Format:    "Audio MP3 (320kbps)",
			FilePath:  "/tmp/sample.mp3",
			FileSize:  5 * 1024 * 1024,
			CreatedAt: time.Now(),
		},
	}

	view := m.View()
	if !strings.Contains(view, "Sample Song") || !strings.Contains(view, "/tmp/sample.mp3") {
		t.Errorf("expected history view to render entry details, got: %s", view)
	}
}

func TestModelDuplicateWarning(t *testing.T) {
	m := InitialModel("", "")
	m.State = StateInputURL
	m.DuplicateHistory = &util.HistoryEntry{
		URL:       "https://youtube.com/watch?v=dup",
		Title:     "Old Video",
		Format:    "Best Video (MP4)",
		FilePath:  "/tmp/old_video.mp4",
		CreatedAt: time.Now(),
	}

	view := m.View()
	if !strings.Contains(view, "ĐÃ TẢI TRƯỚC ĐÂY") {
		t.Errorf("expected duplicate warning in view, got: %s", view)
	}
	if !strings.Contains(view, "Old Video") {
		t.Errorf("expected duplicate title in view, got: %s", view)
	}
}
