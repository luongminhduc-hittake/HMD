package downloader

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"hmd/internal/spotify"
	"hmd/internal/util"
)

// DownloadOptions defines configuration for the download job.
type DownloadOptions struct {
	YtDlpPath           string
	FFmpegPath          string
	GalleryDlPath       string
	URL                 string
	Title               string
	OutputDir           string
	CustomFilename      string
	Preset              FormatPreset
	DownloadPlaylist    bool
	TrimRange           string
	EnableSubtitles     bool
	CookiesBrowser      string
	ConcurrentFragments int
	ConcurrentTracks    int
	MaxRetries          int
	FragmentRetries     int
	IsSpotify           bool
	IsImage             bool
	SpotifyTracks       []spotify.TrackInfo
	OnProgress          func(ProgressUpdate)
	OnStatus            func(string)
}

var (
	playlistItemRegex = regexp.MustCompile(`\[download\]\s+Downloading\s+item\s+(\d+)\s+of\s+(\d+)`)
	errorItemRegex    = regexp.MustCompile(`ERROR:\s*\[[^\]]+\]\s*([^:\s]+):`)
)

// BuildDownloadArgs constructs the CLI arguments for yt-dlp.
func BuildDownloadArgs(opts DownloadOptions) []string {
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = util.GetDefaultDownloadDir()
	}
	outputDir = filepath.Clean(outputDir)

	cf := opts.ConcurrentFragments
	if cf <= 0 {
		cf = 4
	}
	if cf > 8 {
		cf = 8
	}

	mr := opts.MaxRetries
	if mr <= 0 {
		mr = 3
	}
	if mr > 10 {
		mr = 10
	}

	fr := opts.FragmentRetries
	if fr <= 0 {
		fr = 10
	}
	if fr > 10 {
		fr = 10
	}

	args := []string{
		"--newline",
		"--no-mtime",
		"--no-config",
		"--windows-filenames",
		"--progress",
		"--no-quiet",
		"--progress-template", "DOWNLOAD_PROGRESS:%(progress._percent_str)s|%(progress._speed_str)s|%(progress._eta_str)s|%(progress._total_bytes_str|progress._total_bytes_estimate_str)s",
		"--print", "after_move:FINAL_PATH:%(filepath)s",
		"--concurrent-fragments", strconv.Itoa(cf),
		"--retries", strconv.Itoa(mr),
		"--fragment-retries", strconv.Itoa(fr),
		"--retry-sleep", "fragment:exp=1:20",
		"--socket-timeout", "15",
	}

	outputPattern := "%(title)s.%(ext)s"
	if opts.CustomFilename != "" {
		outputPattern = util.SanitizeFilename(opts.CustomFilename) + ".%(ext)s"
	}
	args = append(args, "-o", filepath.Join(outputDir, outputPattern))

	if opts.CookiesBrowser != "" && util.IsAllowedBrowser(opts.CookiesBrowser) {
		args = append(args, "--cookies-from-browser", opts.CookiesBrowser)
	}

	if opts.FFmpegPath != "" {
		ffmpegDir := filepath.Dir(opts.FFmpegPath)
		args = append(args, "--ffmpeg-location", ffmpegDir)
	}

	if opts.DownloadPlaylist {
		args = append(args,
			"--yes-playlist",
			"--ignore-errors",
			"--no-abort-on-error",
			"--skip-playlist-after-errors", "infinite",
		)
	} else {
		args = append(args, "--no-playlist")
	}

	// Apply presets
	switch opts.Preset {
	case PresetBestVideo:
		args = append(args,
			"-f", "bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4] / bv*+ba/b",
			"--merge-output-format", "mp4",
		)
	case Preset1080p:
		args = append(args,
			"-f", "bv*[height<=1080][ext=mp4]+ba[ext=m4a]/b[height<=1080][ext=mp4] / bv*[height<=1080]+ba/b[height<=1080]",
			"--merge-output-format", "mp4",
		)
	case Preset720p:
		args = append(args,
			"-f", "bv*[height<=720][ext=mp4]+ba[ext=m4a]/b[height<=720][ext=mp4] / bv*[height<=720]+ba/b[height<=720]",
			"--merge-output-format", "mp4",
		)
	case PresetAudioMP3, PresetAudioM4A, PresetAudioFLAC, PresetAudioOPUS:
		if opts.Preset == PresetAudioM4A {
			args = append(args, "-f", "ba[ext=m4a]/ba")
		}
		args = append(args, "-x", "--audio-format", string(opts.Preset))
		if opts.Preset == PresetAudioMP3 {
			args = append(args, "--audio-quality", "320K")
		}
		args = append(args, "--embed-thumbnail", "--add-metadata")
	case PresetAudioWAV:
		args = append(args, "-x", "--audio-format", "wav", "--add-metadata")
	case PresetThumbnail:
		args = append(args, "--skip-download", "--write-thumbnail", "--convert-thumbnails", "jpg")
	default:
		args = append(args,
			"-f", "bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4] / bv*+ba/b",
			"--merge-output-format", "mp4",
		)
	}

	// Soft subtitles
	if opts.EnableSubtitles {
		args = append(args,
			"--write-subs",
			"--write-auto-subs",
			"--sub-langs", "vi.*,en.*,vi,en",
			"--embed-subs",
		)
	}

	// Time-range trim
	if strings.TrimSpace(opts.TrimRange) != "" {
		trim := strings.TrimSpace(opts.TrimRange)
		if !strings.HasPrefix(trim, "*") {
			trim = "*" + trim
		}
		args = append(args, "--download-sections", trim, "--force-keyframes-at-cuts")
	}

	if opts.URL != "" {
		args = append(args, "--", opts.URL)
	}

	return args
}

// ExecuteDownload initiates download for given options.
func ExecuteDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.IsSpotify && len(opts.SpotifyTracks) > 0 {
		return runSpotifyDownload(ctx, opts)
	}
	if opts.Preset == PresetImageOriginal || opts.IsImage {
		return executeGalleryDlDownload(ctx, opts)
	}
	return runStandardDownload(ctx, opts)
}
