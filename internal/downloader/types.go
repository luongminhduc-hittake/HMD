package downloader

// FormatPreset represents a download quality and format preset.
type FormatPreset string

const (
	PresetBestVideo FormatPreset = "best"
	Preset1080p     FormatPreset = "1080p"
	Preset720p      FormatPreset = "720p"
	PresetAudioMP3  FormatPreset = "mp3"
	PresetAudioM4A  FormatPreset = "m4a"
)

// PresetOption holds UI presentation metadata for presets.
type PresetOption struct {
	ID          FormatPreset
	Title       string
	Description string
}

// AvailablePresets provides the list of choices for the UI.
var AvailablePresets = []PresetOption{
	{
		ID:          PresetBestVideo,
		Title:       "Best Video (MP4)",
		Description: "Chất lượng cao nhất (4K/2K/1080p) tự động gộp audio",
	},
	{
		ID:          Preset1080p,
		Title:       "Full HD 1080p (MP4)",
		Description: "Độ phân giải 1080p chuẩn, tương thích mọi thiết bị",
	},
	{
		ID:          Preset720p,
		Title:       "HD 720p (MP4)",
		Description: "Độ phân giải 720p tiết kiệm dung lượng",
	},
	{
		ID:          PresetAudioMP3,
		Title:       "Audio MP3 (320kbps)",
		Description: "Âm thanh chất lượng cao, tự động nhúng bìa & ID3 tag",
	},
	{
		ID:          PresetAudioM4A,
		Title:       "Audio M4A gốc (AAC)",
		Description: "Tải stream AAC trực tiếp siêu tốc, không cần convert",
	},
}

// MediaInfo contains inspected metadata of the URL.
type MediaInfo struct {
	Title         string
	Duration      string
	Uploader      string
	IsPlaylist    bool
	PlaylistCount int
	RawURL        string
}

// ProgressUpdate contains real-time progress information.
type ProgressUpdate struct {
	Percent       float64
	Speed         string
	ETA           string
	TotalSize     string
	StatusMessage string
}

// DownloadResult contains final results after a download completes.
type DownloadResult struct {
	FilePath      string
	FileName      string
	FileSize      int64
	FormattedSize string
	Title         string
}
