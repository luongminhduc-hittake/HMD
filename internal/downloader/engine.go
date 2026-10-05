package downloader

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"hmd/internal/util"
)

// DownloadOptions defines configuration for the download job.
type DownloadOptions struct {
	YtDlpPath        string
	FFmpegPath       string
	URL              string
	OutputDir        string
	Preset           FormatPreset
	DownloadPlaylist bool
	TrimRange        string
	EnableSubtitles  bool
	OnProgress       func(ProgressUpdate)
	OnStatus         func(string)
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

	args := []string{
		"--newline",
		"--no-mtime",
		"--progress",
		"--no-quiet",
		"--progress-template", "DOWNLOAD_PROGRESS:%(progress._percent_str)s|%(progress._speed_str)s|%(progress._eta_str)s|%(progress._total_bytes_str|progress._total_bytes_estimate_str)s",
		"--print", "after_move:FINAL_PATH:%(filepath)s",
		"-o", filepath.Join(outputDir, "%(title)s.%(ext)s"),
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
		args = append(args, opts.URL)
	}

	return args
}

// ExecuteDownload initiates yt-dlp with appropriate flags and streams progress updates.
func ExecuteDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.OutputDir == "" {
		opts.OutputDir = util.GetDefaultDownloadDir()
	}
	_ = os.MkdirAll(opts.OutputDir, 0755)

	args := BuildDownloadArgs(opts)

	cmd := exec.CommandContext(ctx, opts.YtDlpPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối stdout: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for complete logging

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("không thể khởi chạy yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	var finalFilePath string
	var lastStatus string
	var lastPercent float64
	downloadedFiles := make(map[string]bool)
	skippedVideos := make(map[string]bool)
	rawSkippedCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "FINAL_PATH:") {
			finalFilePath = strings.TrimPrefix(line, "FINAL_PATH:")
			downloadedFiles[finalFilePath] = true
			continue
		}

		if strings.HasPrefix(line, "[download] Destination: ") {
			finalFilePath = strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination: "))
			downloadedFiles[finalFilePath] = true
		} else if strings.Contains(line, " has already been downloaded") && strings.HasPrefix(line, "[download] ") {
			f := strings.TrimPrefix(line, "[download] ")
			f = strings.TrimSuffix(f, " has already been downloaded")
			finalFilePath = strings.TrimSpace(f)
			downloadedFiles[finalFilePath] = true
		} else if strings.HasPrefix(line, "[Merger] Merging formats into \"") {
			f := strings.TrimPrefix(line, "[Merger] Merging formats into \"")
			f = strings.TrimSuffix(f, "\"")
			finalFilePath = strings.TrimSpace(f)
			downloadedFiles[finalFilePath] = true
		}

		if pMatches := playlistItemRegex.FindStringSubmatch(line); len(pMatches) == 3 {
			if opts.OnStatus != nil {
				opts.OnStatus(fmt.Sprintf("Đang xử lý video %s/%s trong danh sách...", pMatches[1], pMatches[2]))
			}
		}

		if strings.Contains(line, "ERROR:") || strings.Contains(line, "Skipping item") || strings.Contains(line, "is unavailable") || strings.Contains(line, "Private video") {
			if eMatches := errorItemRegex.FindStringSubmatch(line); len(eMatches) == 2 {
				skippedVideos[eMatches[1]] = true
			} else if strings.Contains(line, "ERROR:") {
				rawSkippedCount++
			}
		}

		if strings.HasPrefix(line, "DOWNLOAD_PROGRESS:") {
			raw := strings.TrimPrefix(line, "DOWNLOAD_PROGRESS:")
			parts := strings.Split(raw, "|")
			if len(parts) >= 4 {
				pctStr := strings.Trim(strings.TrimSpace(parts[0]), "%")
				pct, err := strconv.ParseFloat(pctStr, 64)
				if err == nil {
					lastPercent = pct
				}
				speed := strings.TrimSpace(parts[1])
				if speed == "Unknown B/s" || speed == "NA" {
					speed = ""
				}
				eta := strings.TrimSpace(parts[2])
				if eta == "Unknown" || eta == "NA" {
					eta = ""
				}
				total := strings.TrimSpace(parts[3])
				if total == "NA" {
					total = ""
				}

				if opts.OnProgress != nil {
					opts.OnProgress(ProgressUpdate{
						Percent:       lastPercent,
						Speed:         speed,
						ETA:           eta,
						TotalSize:     total,
						StatusMessage: "Đang tải dữ liệu...",
					})
				}
			}
			continue
		}

		// Check postprocessing stages
		statusMsg := ""
		switch {
		case strings.Contains(line, "[ExtractAudio]"):
			statusMsg = "Đang chuyển đổi định dạng âm thanh..."
		case strings.Contains(line, "[Merger]"):
			statusMsg = "Đang hợp nhất video và audio..."
		case strings.Contains(line, "[EmbedThumbnail]"):
			statusMsg = "Đang nhúng ảnh bìa thumbnail..."
		case strings.Contains(line, "[Metadata]") || strings.Contains(line, "[Fixup]"):
			statusMsg = "Đang hoàn thiện siêu dữ liệu..."
		case strings.Contains(line, "Deleting original file"):
			statusMsg = "Đang dọn dẹp các tệp tạm..."
		}

		if statusMsg != "" && statusMsg != lastStatus {
			lastStatus = statusMsg
			if opts.OnStatus != nil {
				opts.OnStatus(statusMsg)
			}
		}
	}

	waitErr := cmd.Wait()
	downloadedCount := len(downloadedFiles)
	skippedCount := len(skippedVideos)
	if skippedCount == 0 && rawSkippedCount > 0 {
		skippedCount = rawSkippedCount
	}

	if waitErr != nil {
		if opts.DownloadPlaylist && downloadedCount > 0 {
			// Playlist completed with some videos downloaded; ignore non-zero exit caused by skipped items
		} else {
			return nil, fmt.Errorf("lỗi trong quá trình tải xuống: %w", waitErr)
		}
	}

	res := &DownloadResult{
		FilePath:        finalFilePath,
		IsPlaylist:      opts.DownloadPlaylist,
		DownloadedCount: downloadedCount,
		SkippedCount:    skippedCount,
	}

	if opts.DownloadPlaylist {
		res.FilePath = opts.OutputDir
		res.FileName = filepath.Base(opts.OutputDir)
	} else if finalFilePath != "" {
		res.FileName = filepath.Base(finalFilePath)
		if fi, err := os.Stat(finalFilePath); err == nil {
			res.FileSize = fi.Size()
			res.FormattedSize = util.FormatBytes(fi.Size())
		}
	} else {
		res.FilePath = opts.OutputDir
	}

	return res, nil
}
