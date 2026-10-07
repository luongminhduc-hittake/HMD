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
	"sync"

	"hmd/internal/spotify"
	"hmd/internal/util"
)

// DownloadOptions defines configuration for the download job.
type DownloadOptions struct {
	YtDlpPath           string
	FFmpegPath          string
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
	SpotifyTracks       []spotify.TrackInfo
	OnProgress          func(ProgressUpdate)
	OnStatus            func(string)
}

var (
	execCommandContext = exec.CommandContext
	playlistItemRegex  = regexp.MustCompile(`\[download\]\s+Downloading\s+item\s+(\d+)\s+of\s+(\d+)`)
	errorItemRegex     = regexp.MustCompile(`ERROR:\s*\[[^\]]+\]\s*([^:\s]+):`)
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
		return executeSpotifyDownload(ctx, opts)
	}
	return executeStandardDownload(ctx, opts)
}

func executeStandardDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.OutputDir == "" {
		opts.OutputDir = util.GetDefaultDownloadDir()
	}
	_ = os.MkdirAll(opts.OutputDir, 0755)

	args := BuildDownloadArgs(opts)

	cmd := execCommandContext(ctx, opts.YtDlpPath, args...)
	prepareCommand(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối stdout: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout for complete logging

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("không thể khởi chạy yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var finalFilePath string
	var lastStatus string
	var lastPercent float64
	var hasCookieLock bool
	downloadedFiles := make(map[string]bool)
	skippedVideos := make(map[string]bool)
	rawSkippedCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, "database is locked") {
			hasCookieLock = true
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
		} else if strings.Contains(line, "Writing video thumbnail") && strings.Contains(line, " to: ") {
			parts := strings.Split(line, " to: ")
			if len(parts) == 2 {
				finalFilePath = strings.TrimSpace(parts[1])
				downloadedFiles[finalFilePath] = true
			}
		} else if strings.Contains(line, "Converting thumbnail") && strings.Contains(line, " to: ") {
			parts := strings.Split(line, " to: ")
			if len(parts) == 2 {
				finalFilePath = strings.Trim(strings.TrimSpace(parts[1]), "\"")
				downloadedFiles[finalFilePath] = true
			}
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
		case strings.Contains(line, "[ThumbnailsConvertor]"), strings.Contains(line, "Writing video thumbnail"):
			statusMsg = "Đang trích xuất ảnh bìa chất lượng cao..."
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

	_ = stdout.Close()
	waitErr := cmd.Wait()
	if scanErr := scanner.Err(); scanErr != nil && waitErr == nil {
		waitErr = fmt.Errorf("lỗi đọc dữ liệu từ yt-dlp: %w", scanErr)
	}

	downloadedCount := len(downloadedFiles)
	skippedCount := len(skippedVideos)
	if skippedCount == 0 && rawSkippedCount > 0 {
		skippedCount = rawSkippedCount
	}

	if opts.Preset == PresetThumbnail && finalFilePath != "" {
		if downloadedCount == 0 {
			downloadedCount = 1
		}
		if opts.OnProgress != nil {
			opts.OnProgress(ProgressUpdate{
				Percent:       100,
				StatusMessage: "Đã tải xong ảnh bìa!",
			})
		}
	}

	if waitErr != nil {
		if hasCookieLock {
			return nil, fmt.Errorf("Trình duyệt đang mở và khóa file cookies. Cậu chủ vui lòng đóng trình duyệt rồi thử lại.")
		}
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

func executeSpotifyDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	// If single track and not downloading as playlist
	if len(opts.SpotifyTracks) == 1 && !opts.DownloadPlaylist {
		track := opts.SpotifyTracks[0]
		singleOpts := opts
		singleOpts.IsSpotify = false
		singleOpts.CustomFilename = fmt.Sprintf("%s - %s", track.Artist, track.Title)
		if opts.OnStatus != nil {
			opts.OnStatus(fmt.Sprintf("Đang tìm & tải: %s - %s", track.Artist, track.Title))
		}

		// Tier 1: ytsearch1:... audio
		singleOpts.URL = fmt.Sprintf("ytsearch1:%s - %s audio", track.Artist, track.Title)
		res, err := executeStandardDownload(ctx, singleOpts)
		if err != nil || res == nil || res.FilePath == "" {
			// Tier 2: ytsearch1:...
		singleOpts.URL = fmt.Sprintf("ytsearch1:%s - %s", track.Artist, track.Title)
			res, err = executeStandardDownload(ctx, singleOpts)
		}
		if err != nil {
			return nil, fmt.Errorf("không thể tìm hoặc tải bài hát %q: %w", track.Title, err)
		}

		// Inject metadata & cover
		_ = InjectMetadata(ctx, opts.FFmpegPath, res.FilePath, track)
		if fi, err := os.Stat(res.FilePath); err == nil {
			res.FileSize = fi.Size()
			res.FormattedSize = util.FormatBytes(fi.Size())
		}
		res.Title = fmt.Sprintf("%s - %s", track.Artist, track.Title)
		return res, nil
	}

	// Multiple tracks / Album / Playlist (Concurrent Worker Pool)
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = util.GetDefaultDownloadDir()
	}
	if opts.Title != "" {
		outputDir = filepath.Join(outputDir, util.SanitizeFilename(opts.Title))
	}
	_ = os.MkdirAll(outputDir, 0755)

	total := len(opts.SpotifyTracks)
	concurrency := opts.ConcurrentTracks
	if concurrency <= 0 {
		concurrency = 3
	}
	if concurrency > 5 {
		concurrency = 5
	}
	if total < concurrency {
		concurrency = total
	}

	var (
		mu              sync.Mutex
		downloadedCount int
		skippedCount    int
		completedCount  int
		failedTracks    []string
		wg              sync.WaitGroup
	)

	jobs := make(chan int, total)
	for i := 0; i < total; i++ {
		jobs <- i
	}
	close(jobs)

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				if ctx.Err() != nil {
					return
				}
				track := opts.SpotifyTracks[idx]

				if !track.IsPlayable {
					mu.Lock()
					skippedCount++
					completedCount++
					failedTracks = append(failedTracks, fmt.Sprintf("%s - %s (bị khóa bản quyền trên Spotify)", track.Artist, track.Title))
					mu.Unlock()
					continue
				}

				mu.Lock()
				if opts.OnStatus != nil {
					opts.OnStatus(fmt.Sprintf("[%d/%d] Đang tải: %s - %s", completedCount+1, total, track.Artist, track.Title))
				}
				mu.Unlock()

				trackOpts := opts
				trackOpts.IsSpotify = false
				trackOpts.OutputDir = outputDir
				trackOpts.DownloadPlaylist = false
				trackOpts.CustomFilename = fmt.Sprintf("%s - %s", track.Artist, track.Title)
				trackOpts.OnProgress = nil

				// Tier 1: ytsearch1:... audio
				trackOpts.URL = fmt.Sprintf("ytsearch1:%s - %s audio", track.Artist, track.Title)
				singleRes, err := executeStandardDownload(ctx, trackOpts)
				if err != nil || singleRes == nil || singleRes.FilePath == "" {
					// Tier 2: ytsearch1:...
					trackOpts.URL = fmt.Sprintf("ytsearch1:%s - %s", track.Artist, track.Title)
					singleRes, err = executeStandardDownload(ctx, trackOpts)
				}

				mu.Lock()
				completedCount++
				if err != nil || singleRes == nil || singleRes.FilePath == "" {
					skippedCount++
					failedTracks = append(failedTracks, fmt.Sprintf("%s - %s (không tìm thấy trên YouTube)", track.Artist, track.Title))
				} else {
					downloadedCount++
					_ = InjectMetadata(ctx, opts.FFmpegPath, singleRes.FilePath, track)
				}
				if opts.OnProgress != nil {
					opts.OnProgress(ProgressUpdate{
						Percent:       float64(completedCount) / float64(total) * 100,
						StatusMessage: fmt.Sprintf("[%d/%d] Đã xong %d bài", completedCount, total, downloadedCount),
					})
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return &DownloadResult{
		FilePath:        outputDir,
		FileName:        filepath.Base(outputDir),
		Title:           opts.Title,
		IsPlaylist:      true,
		DownloadedCount: downloadedCount,
		SkippedCount:    skippedCount,
		FailedTracks:    failedTracks,
	}, nil
}
