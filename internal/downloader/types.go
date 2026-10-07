package downloader

import "hmd/internal/spotify"

// FormatPreset represents a download quality and format preset.
type FormatPreset string

const (
	PresetBestVideo FormatPreset = "best"
	Preset1080p     FormatPreset = "1080p"
	Preset720p      FormatPreset = "720p"
	PresetAudioMP3  FormatPreset = "mp3"
	PresetAudioM4A  FormatPreset = "m4a"
	PresetAudioWAV  FormatPreset = "wav"
	PresetAudioFLAC FormatPreset = "flac"
	PresetAudioOPUS FormatPreset = "opus"
	PresetThumbnail     FormatPreset = "thumb"
	PresetImageOriginal FormatPreset = "image"
)

// MediaType represents the detected media kind.
type MediaType string

const (
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeImage MediaType = "image"
)

// PresetOption holds UI presentation metadata for presets.
type PresetOption struct {
	ID          FormatPreset
	Title       string
	Description string
}

var (
	PresetOptBestVideo = PresetOption{
		ID:          PresetBestVideo,
		Title:       "Best Video (MP4)",
		Description: "Chất lượng cao nhất tự động gộp audio",
	}
	PresetOpt1080p = PresetOption{
		ID:          Preset1080p,
		Title:       "Full HD 1080p (MP4)",
		Description: "Độ phân giải 1080p chuẩn, tương thích mọi thiết bị",
	}
	PresetOpt720p = PresetOption{
		ID:          Preset720p,
		Title:       "HD 720p (MP4)",
		Description: "Độ phân giải 720p tiết kiệm dung lượng",
	}
	PresetOptAudioMP3 = PresetOption{
		ID:          PresetAudioMP3,
		Title:       "Audio MP3 (320kbps)",
		Description: "Âm thanh chất lượng cao, tự động nhúng bìa & ID3 tag",
	}
	PresetOptAudioM4A = PresetOption{
		ID:          PresetAudioM4A,
		Title:       "Audio M4A gốc (AAC)",
		Description: "Tải stream AAC trực tiếp siêu tốc, không cần convert",
	}
	PresetOptAudioWAV = PresetOption{
		ID:          PresetAudioWAV,
		Title:       "Audio WAV (Lossless)",
		Description: "Âm thanh nguyên bản phòng thu PCM, không nén",
	}
	PresetOptAudioFLAC = PresetOption{
		ID:          PresetAudioFLAC,
		Title:       "Audio FLAC (Lossless)",
		Description: "Âm thanh nguyên bản không nén chất lượng phòng thu, nhúng bìa & ID3",
	}
	PresetOptAudioOPUS = PresetOption{
		ID:          PresetAudioOPUS,
		Title:       "Audio OPUS (Chất lượng cao)",
		Description: "Chuẩn nén thế hệ mới, tối ưu băng thông & giữ nguyên chi tiết",
	}
	PresetOptThumbnail = PresetOption{
		ID:          PresetThumbnail,
		Title:       "Thumbnail (Ảnh bìa)",
		Description: "Tải riêng ảnh bìa sắc nét (JPG), không tải video",
	}
	PresetOptImageOriginal = PresetOption{
		ID:          PresetImageOriginal,
		Title:       "Ảnh gốc (High-Res)",
		Description: "Tải ảnh gốc / toàn bộ gallery chất lượng cao nhất",
	}
)

// AudioPresets contains presets suitable for audio-only sources.
var AudioPresets = []PresetOption{
	PresetOptAudioMP3,
	PresetOptAudioM4A,
	PresetOptAudioWAV,
	PresetOptAudioFLAC,
	PresetOptAudioOPUS,
	PresetOptThumbnail,
}

// ImagePresets contains presets suitable for photo/gallery sources.
var ImagePresets = []PresetOption{
	PresetOptImageOriginal,
}

// AvailablePresets provides the fallback list of choices for the UI.
var AvailablePresets = []PresetOption{
	PresetOptBestVideo,
	PresetOpt1080p,
	PresetOpt720p,
	PresetOptAudioMP3,
	PresetOptAudioM4A,
	PresetOptAudioWAV,
	PresetOptAudioFLAC,
	PresetOptAudioOPUS,
	PresetOptThumbnail,
}

// DetermineAvailablePresets returns the list of available preset options based on media type and video properties.
func DetermineAvailablePresets(mediaType MediaType, maxHeight int) []PresetOption {
	switch mediaType {
	case MediaTypeImage:
		return ImagePresets
	case MediaTypeAudio:
		return AudioPresets
	default:
		var opts []PresetOption
		opts = append(opts, PresetOptBestVideo)
		if maxHeight >= 1080 || maxHeight == 0 {
			opts = append(opts, PresetOpt1080p)
		}
		if maxHeight >= 720 || maxHeight == 0 {
			opts = append(opts, PresetOpt720p)
		}
		opts = append(opts, AudioPresets...)
		return opts
	}
}

// MediaInfo contains inspected metadata of the URL.
type MediaInfo struct {
	Title            string
	Duration         string
	Uploader         string
	IsPlaylist       bool
	RawURL           string
	IsSpotify        bool
	IsImage          bool
	SpotifyTracks    []spotify.TrackInfo
	ThumbnailURL     string
	MediaType        MediaType
	MaxHeight        int
	AvailablePresets []PresetOption
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
	FilePath        string
	FileName        string
	FileSize        int64
	FormattedSize   string
	Title           string
	IsPlaylist      bool
	DownloadedCount int
	SkippedCount    int
	FailedTracks    []string
}
