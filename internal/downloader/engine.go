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

	"ytdownloader/internal/util"
)

// DownloadOptions defines configuration for the download job.
type DownloadOptions struct {
	YtDlpPath        string
	FFmpegPath       string
	URL              string
	OutputDir        string
	Preset           FormatPreset
	DownloadPlaylist bool
	OnProgress       func(ProgressUpdate)
	OnStatus         func(string)
}

var fallbackRegex = regexp.MustCompile(`\[download\]\s+(\d+(?:\.\d+)?)%\s+of\s+~?(\S+)\s+at\s+(\S+)\s+ETA\s+(\S+)`)

// ExecuteDownload initiates yt-dlp with appropriate flags and streams progress updates.
func ExecuteDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.OutputDir == "" {
		opts.OutputDir = util.GetDefaultDownloadDir()
	}
	_ = os.MkdirAll(opts.OutputDir, 0755)

	args := []string{
		"--newline",
		"--no-mtime",
		"--progress",
		"--no-quiet",
		"--progress-template", "DOWNLOAD_PROGRESS:%(progress._percent_str)s|%(progress._speed_str)s|%(progress._eta_str)s|%(progress._total_bytes_str|progress._total_bytes_estimate_str)s",
		"--print", "after_move:FINAL_PATH:%(filepath)s",
		"-o", filepath.Join(opts.OutputDir, "%(title)s.%(ext)s"),
	}

	if opts.FFmpegPath != "" {
		ffmpegDir := filepath.Dir(opts.FFmpegPath)
		args = append(args, "--ffmpeg-location", ffmpegDir)
	}

	if opts.DownloadPlaylist {
		args = append(args, "--yes-playlist")
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
	case PresetAudioMP3:
		args = append(args,
			"-x",
			"--audio-format", "mp3",
			"--audio-quality", "320K",
			"--embed-thumbnail",
			"--add-metadata",
		)
	case PresetAudioM4A:
		args = append(args,
			"-f", "ba[ext=m4a]/ba",
			"-x",
			"--audio-format", "m4a",
			"--embed-thumbnail",
			"--add-metadata",
		)
	default:
		args = append(args,
			"-f", "bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4] / bv*+ba/b",
			"--merge-output-format", "mp4",
		)
	}

	args = append(args, opts.URL)

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

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "FINAL_PATH:") {
			finalFilePath = strings.TrimPrefix(line, "FINAL_PATH:")
			continue
		}

		if strings.HasPrefix(line, "[download] Destination: ") {
			finalFilePath = strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination: "))
		} else if strings.Contains(line, " has already been downloaded") && strings.HasPrefix(line, "[download] ") {
			f := strings.TrimPrefix(line, "[download] ")
			f = strings.TrimSuffix(f, " has already been downloaded")
			finalFilePath = strings.TrimSpace(f)
		} else if strings.HasPrefix(line, "[Merger] Merging formats into \"") {
			f := strings.TrimPrefix(line, "[Merger] Merging formats into \"")
			f = strings.TrimSuffix(f, "\"")
			finalFilePath = strings.TrimSpace(f)
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

		// Fallback regex matching
		if matches := fallbackRegex.FindStringSubmatch(line); len(matches) == 5 {
			pct, _ := strconv.ParseFloat(matches[1], 64)
			lastPercent = pct
			eta := matches[4]
			if eta == "Unknown" || eta == "NA" {
				eta = ""
			}
			speed := matches[3]
			if speed == "Unknown B/s" || speed == "NA" {
				speed = ""
			}
			if opts.OnProgress != nil {
				opts.OnProgress(ProgressUpdate{
					Percent:       lastPercent,
					TotalSize:     matches[2],
					Speed:         speed,
					ETA:           eta,
					StatusMessage: "Đang tải dữ liệu...",
				})
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

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("lỗi trong quá trình tải xuống: %w", err)
	}

	res := &DownloadResult{
		FilePath: finalFilePath,
	}

	if finalFilePath != "" {
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
