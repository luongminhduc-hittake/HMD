package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hmd/internal/util"
)

// YtDlpParser parses lines streamed from yt-dlp to track download progress,
// identify downloaded file paths, and detect status changes.
type YtDlpParser struct {
	opts            DownloadOptions
	finalFilePath   string
	downloadedFiles map[string]bool
	skippedVideos   map[string]bool
	rawSkippedCount int
	hasCookieLock   bool
	lastPercent     float64
}

// NewYtDlpParser creates an initialized parser for yt-dlp stdout.
func NewYtDlpParser(opts DownloadOptions) *YtDlpParser {
	return &YtDlpParser{
		opts:            opts,
		downloadedFiles: make(map[string]bool),
		skippedVideos:   make(map[string]bool),
	}
}

// FeedLine processes a single output line from yt-dlp.
func (p *YtDlpParser) FeedLine(line string) {
	// HOT PATH: Check download progress first without regex or heavy checks
	if strings.HasPrefix(line, "DOWNLOAD_PROGRESS:") {
		p.parseProgress(line[len("DOWNLOAD_PROGRESS:"):])
		return
	}

	// Output paths
	if strings.HasPrefix(line, "FINAL_PATH:") {
		path := strings.TrimPrefix(line, "FINAL_PATH:")
		p.finalFilePath = path
		p.downloadedFiles[path] = true
		return
	}

	if strings.HasPrefix(line, "[download] Destination: ") {
		path := strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination: "))
		p.finalFilePath = path
		p.downloadedFiles[path] = true
	} else if strings.Contains(line, " has already been downloaded") && strings.HasPrefix(line, "[download] ") {
		f := strings.TrimPrefix(line, "[download] ")
		f = strings.TrimSuffix(f, " has already been downloaded")
		p.finalFilePath = strings.TrimSpace(f)
		p.downloadedFiles[p.finalFilePath] = true
	} else if strings.HasPrefix(line, "[Merger] Merging formats into \"") {
		f := strings.TrimPrefix(line, "[Merger] Merging formats into \"")
		f = strings.TrimSuffix(f, "\"")
		p.finalFilePath = strings.TrimSpace(f)
		p.downloadedFiles[p.finalFilePath] = true
	} else if strings.Contains(line, "Writing video thumbnail") && strings.Contains(line, " to: ") {
		parts := strings.Split(line, " to: ")
		if len(parts) == 2 {
			p.finalFilePath = strings.TrimSpace(parts[1])
			p.downloadedFiles[p.finalFilePath] = true
		}
	} else if strings.Contains(line, "Converting thumbnail") && strings.Contains(line, " to: ") {
		parts := strings.Split(line, " to: ")
		if len(parts) == 2 {
			p.finalFilePath = strings.Trim(strings.TrimSpace(parts[1]), "\"")
			p.downloadedFiles[p.finalFilePath] = true
		}
	}

	// Playlist progress
	if strings.HasPrefix(line, "[download] Downloading item ") {
		if pMatches := playlistItemRegex.FindStringSubmatch(line); len(pMatches) == 3 {
			if p.opts.OnStatus != nil {
				p.opts.OnStatus(fmt.Sprintf("Đang xử lý video %s/%s trong danh sách...", pMatches[1], pMatches[2]))
			}
		}
	}

	// Error & skip detection
	if strings.Contains(line, "database is locked") {
		p.hasCookieLock = true
	}
	if strings.Contains(line, "ERROR:") || strings.Contains(line, "Skipping item") || strings.Contains(line, "is unavailable") || strings.Contains(line, "Private video") {
		if eMatches := errorItemRegex.FindStringSubmatch(line); len(eMatches) == 2 {
			p.skippedVideos[eMatches[1]] = true
		} else if strings.Contains(line, "ERROR:") {
			p.rawSkippedCount++
		}
	}

	// Post-processing stages
	p.checkPostprocessing(line)
}

func (p *YtDlpParser) parseProgress(raw string) {
	parts := strings.Split(raw, "|")
	if len(parts) < 4 {
		return
	}

	pctStr := strings.Trim(strings.TrimSpace(parts[0]), "%")
	if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
		p.lastPercent = pct
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

	if p.opts.OnProgress != nil {
		p.opts.OnProgress(ProgressUpdate{
			Percent:       p.lastPercent,
			Speed:         speed,
			ETA:           eta,
			TotalSize:     total,
			StatusMessage: "Đang tải dữ liệu...",
		})
	}
}

func (p *YtDlpParser) checkPostprocessing(line string) {
	if p.opts.OnStatus == nil {
		return
	}

	var statusMsg string
	switch {
	case strings.Contains(line, "[ExtractAudio]"):
		statusMsg = "Đang chuyển đổi định dạng âm thanh..."
	case strings.Contains(line, "[Merger]"):
		statusMsg = "Đang hợp nhất video và audio..."
	case strings.Contains(line, "[EmbedThumbnail]"):
		statusMsg = "Đang nhúng ảnh bìa thumbnail..."
	case strings.Contains(line, "[ThumbnailsConvertor]"), strings.Contains(line, "Writing video thumbnail"):
		statusMsg = "Đang xử lý ảnh bìa thumbnail..."
	case strings.Contains(line, "[FixupM4a]"):
		statusMsg = "Đang tối ưu cấu trúc container M4A..."
	case strings.Contains(line, "[Metadata]"):
		statusMsg = "Đang cập nhật thẻ thông tin metadata..."
	}

	if statusMsg != "" {
		p.opts.OnStatus(statusMsg)
	}
}

// Result constructs the final DownloadResult from collected state.
func (p *YtDlpParser) Result() (*DownloadResult, error) {
	var fileSize int64
	formattedSize := "N/A"
	if p.finalFilePath != "" {
		if fi, err := os.Stat(p.finalFilePath); err == nil {
			fileSize = fi.Size()
			formattedSize = util.FormatBytes(fileSize)
		}
	}

	if p.opts.OnProgress != nil && p.opts.Preset == PresetThumbnail {
		p.opts.OnProgress(ProgressUpdate{
			Percent:       100,
			TotalSize:     formattedSize,
			StatusMessage: "Đã tải xong ảnh bìa!",
		})
	}

	downloadedCount := len(p.downloadedFiles)
	if !p.opts.DownloadPlaylist {
		downloadedCount = 1
	}

	skippedCount := len(p.skippedVideos)
	if skippedCount == 0 && p.rawSkippedCount > 0 {
		skippedCount = p.rawSkippedCount
	}

	return &DownloadResult{
		FilePath:        p.finalFilePath,
		FileName:        filepath.Base(p.finalFilePath),
		FileSize:        fileSize,
		FormattedSize:   formattedSize,
		Title:           p.opts.Title,
		IsPlaylist:      p.opts.DownloadPlaylist,
		DownloadedCount: downloadedCount,
		SkippedCount:    skippedCount,
	}, nil
}
