package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/downloader"
	"hmd/internal/spotify"
	"hmd/internal/util"
)

func TestModelTrimAndSubtitles(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
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
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
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
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
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
	if !strings.Contains(view, "Ctrl+P: Mở tệp cũ") {
		t.Errorf("expected view to show Ctrl+P shortcut hint, got: %s", view)
	}
}

func TestModelPresetsSelection(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateSelectPreset

	// Select preset 6 (WAV)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	model2 := m2.(Model)
	if model2.PresetIndex != 5 {
		t.Errorf("expected PresetIndex 5 for key '6', got %d", model2.PresetIndex)
	}
	if downloader.AvailablePresets[model2.PresetIndex].ID != downloader.PresetAudioWAV {
		t.Errorf("expected PresetAudioWAV, got %s", downloader.AvailablePresets[model2.PresetIndex].ID)
	}

	// Select preset 9 (Thumbnail)
	m3, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	model3 := m3.(Model)
	if model3.PresetIndex != 8 {
		t.Errorf("expected PresetIndex 8 for key '9', got %d", model3.PresetIndex)
	}
	if downloader.AvailablePresets[model3.PresetIndex].ID != downloader.PresetThumbnail {
		t.Errorf("expected PresetThumbnail, got %s", downloader.AvailablePresets[model3.PresetIndex].ID)
	}
}

func TestModelCookiesToggle(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateInputURL
	m.Input.SetValue("")

	// Press ctrl+b to cycle to chrome
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	model2 := m2.(Model)
	if model2.CookiesBrowser != "chrome" {
		t.Errorf("expected cookies chrome, got %s", model2.CookiesBrowser)
	}
	if !strings.Contains(model2.View(), "CHROME") {
		t.Errorf("expected view to contain CHROME cookies badge, got: %s", model2.View())
	}

	// Press ctrl+b to cycle to firefox
	m3, _ := model2.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	model3 := m3.(Model)
	if model3.CookiesBrowser != "firefox" {
		t.Errorf("expected cookies firefox, got %s", model3.CookiesBrowser)
	}
}

func TestModelEscClearInput(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateInputURL
	m.Input.SetValue("https://youtube.com/watch?v=abc")
	m.DuplicateHistory = &util.HistoryEntry{URL: "https://youtube.com/watch?v=abc"}

	// First Esc: clears input and duplicate history
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model2 := m2.(Model)
	if model2.Input.Value() != "" {
		t.Errorf("expected input to be cleared, got %q", model2.Input.Value())
	}
	if model2.DuplicateHistory != nil {
		t.Errorf("expected DuplicateHistory to be nil after esc")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd on input clear, got %v", cmd)
	}

	// Second Esc when already empty: quits
	_, cmd2 := model2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd2 == nil {
		t.Errorf("expected quit cmd on empty input esc, got nil")
	}
}

func TestModelErrorActions(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateError
	m.ErrorMessage = "Test Error"
	m.ActionIndex = 0

	view := m.View()
	if !strings.Contains(view, "Thử lại với link khác") ||
		!strings.Contains(view, "Cập nhật yt-dlp mới nhất & thử lại") ||
		!strings.Contains(view, "Thoát chương trình") {
		t.Errorf("expected viewError to render 3 actions, got: %s", view)
	}

	// Navigate down to item 1
	m1, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model1 := m1.(Model)
	if model1.ActionIndex != 1 {
		t.Errorf("expected ActionIndex 1, got %d", model1.ActionIndex)
	}

	// Navigate down to item 2
	m2, _ := model1.Update(tea.KeyMsg{Type: tea.KeyDown})
	model2 := m2.(Model)
	if model2.ActionIndex != 2 {
		t.Errorf("expected ActionIndex 2, got %d", model2.ActionIndex)
	}

	// Cannot navigate beyond item 2
	m2b, _ := model2.Update(tea.KeyMsg{Type: tea.KeyDown})
	model2b := m2b.(Model)
	if model2b.ActionIndex != 2 {
		t.Errorf("expected ActionIndex to stay 2, got %d", model2b.ActionIndex)
	}

	// Select action 1 (Update yt-dlp) via Enter
	m3, cmd3 := model1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model3 := m3.(Model)
	if model3.State != StateUpdating {
		t.Errorf("expected StateUpdating after selecting action 1, got %v", model3.State)
	}
	if cmd3 == nil {
		t.Errorf("expected non-nil cmd for updating yt-dlp")
	}

	// Select action 0 (Retry) via Enter
	m0, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model0 := m0.(Model)
	if model0.State != StateInputURL {
		t.Errorf("expected StateInputURL after selecting action 0, got %v", model0.State)
	}
}

func TestModelCtrlYUpdate(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateInputURL

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	model2 := m2.(Model)
	if model2.State != StateUpdating {
		t.Errorf("expected StateUpdating on Ctrl+Y, got %v", model2.State)
	}
	if cmd == nil {
		t.Errorf("expected update cmd on Ctrl+Y, got nil")
	}
}

func TestModelNoBareLetterShortcuts(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateInputURL
	m.Input.SetValue("")

	// Pressing 'b', 'c', 'h', 'u', 'o' should NOT trigger shortcuts; should type into input
	for _, r := range []rune{'b', 'c', 'h', 'u', 'o'} {
		mRes, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mRes.(Model)
		if m.State != StateInputURL {
			t.Errorf("expected StateInputURL to remain for key %c, got %v", r, m.State)
		}
	}

	if m.Input.Value() != "bchuo" {
		t.Errorf("expected input to contain 'bchuo', got %q", m.Input.Value())
	}
}

func TestModelPlatformHints(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateInputURL
	m.Input.SetValue("")

	view := m.View()
	if !strings.Contains(view, "YouTube") || !strings.Contains(view, "Spotify") || !strings.Contains(view, "Reddit") {
		t.Errorf("expected platform list in view, got: %s", view)
	}
	if !strings.Contains(view, "Cứ thử dán link vào xem tải được không nhé") {
		t.Errorf("expected fallback hint in view, got: %s", view)
	}

	// When entering an unknown media link
	m.Input.SetValue("https://customdomain.tv/video/123")
	view2 := m.View()
	if !strings.Contains(view2, "cứ thử tải xem được không nhé") {
		t.Errorf("expected fallback try prompt for unknown media, got: %s", view2)
	}
}

func TestModelSpotifyPresetAndViews(t *testing.T) {
	t.Setenv("HMD_CONFIG_DIR", t.TempDir())
	m := InitialModel("", "")
	m.State = StateSelectPreset
	m.MediaInfo = &downloader.MediaInfo{
		Title:     "Never Gonna Give You Up",
		Uploader:  "Rick Astley",
		Duration:  "03:33",
		IsSpotify: true,
	}

	// Spotify track view shows note icon and artist
	view := m.View()
	if !strings.Contains(view, "🎵") || !strings.Contains(view, "Rick Astley") {
		t.Errorf("expected Spotify note icon and artist in preset view, got: %s", view)
	}

	// Test playlist view for Spotify
	m.State = StateSelectPlaylist
	m.MediaInfo.IsPlaylist = true
	m.MediaInfo.SpotifyTracks = []spotify.TrackInfo{
		{Title: "Song 1", Artist: "Artist 1"},
		{Title: "Song 2", Artist: "Artist 2"},
	}
	plView := m.View()
	if !strings.Contains(plView, "SPOTIFY") || !strings.Contains(plView, "2 bài hát") {
		t.Errorf("expected Spotify playlist view, got: %s", plView)
	}

	// Test completed view with failed tracks
	m.State = StateCompleted
	m.Result = &downloader.DownloadResult{
		IsPlaylist:      true,
		Title:           "Test Spotify Album",
		DownloadedCount: 1,
		SkippedCount:    1,
		FailedTracks:    []string{"Blocked Song (bị khóa bản quyền)"},
	}
	compView := m.View()
	if !strings.Contains(compView, "Blocked Song (bị khóa bản quyền)") {
		t.Errorf("expected FailedTracks in completed view, got: %s", compView)
	}
}




