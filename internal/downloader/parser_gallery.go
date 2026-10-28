package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hmd/internal/util"
)

// GalleryParser parses stdout lines from gallery-dl to identify downloaded files.
type GalleryParser struct {
	opts            DownloadOptions
	finalFilePath   string
	downloadedFiles map[string]bool
}

// NewGalleryParser creates an initialized parser for gallery-dl stdout.
func NewGalleryParser(opts DownloadOptions) *GalleryParser {
	return &GalleryParser{
		opts:            opts,
		downloadedFiles: make(map[string]bool),
	}
}

// FeedLine processes a single line of gallery-dl output.
func (p *GalleryParser) FeedLine(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}

	// gallery-dl prefixes "# " when a file already exists on disk and was not re-downloaded.
	if strings.HasPrefix(trimmed, "# ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
	}

	// Match paths output by gallery-dl
	if strings.HasPrefix(trimmed, p.opts.OutputDir) ||
		strings.HasPrefix(trimmed, "/") ||
		(len(trimmed) > 2 && trimmed[1] == ':') {
		p.finalFilePath = trimmed
		p.downloadedFiles[trimmed] = true
		if p.opts.OnStatus != nil {
			p.opts.OnStatus(fmt.Sprintf("Đã tải: %s", filepath.Base(trimmed)))
		}
	}
}

// Result constructs the final DownloadResult from gallery-dl state.
func (p *GalleryParser) Result() (*DownloadResult, error) {
	count := len(p.downloadedFiles)
	if count == 0 || p.finalFilePath == "" {
		// gallery-dl exiting 0 without printing a path means nothing was downloaded
		// (login wall, blocked request, empty gallery). Reporting success here would
		// show the user a "completed" screen with no file.
		return nil, fmt.Errorf("gallery-dl không tải được tệp nào (trang bị chặn, cần đăng nhập, hoặc gallery trống)")
	}

	var fileSize int64
	formattedSize := "N/A"
	if p.finalFilePath != "" {
		if fi, err := os.Stat(p.finalFilePath); err == nil {
			fileSize = fi.Size()
			formattedSize = util.FormatBytes(fileSize)
		}
	}

	if p.opts.OnProgress != nil {
		p.opts.OnProgress(ProgressUpdate{
			Percent:       100,
			TotalSize:     formattedSize,
			StatusMessage: "Tải hoàn tất!",
		})
	}

	return &DownloadResult{
		FilePath:        p.finalFilePath,
		FileName:        filepath.Base(p.finalFilePath),
		FileSize:        fileSize,
		FormattedSize:   formattedSize,
		Title:           p.opts.Title,
		DownloadedCount: count,
	}, nil
}
