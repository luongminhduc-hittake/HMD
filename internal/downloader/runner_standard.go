package downloader

import (
	"context"
	"fmt"
)

// runStandardDownload handles yt-dlp execution for single videos and playlists.
func runStandardDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	binPath := opts.YtDlpPath
	if binPath == "" {
		binPath = "yt-dlp"
	}

	args := BuildDownloadArgs(opts)
	parser := NewYtDlpParser(opts)

	err := streamCommand(ctx, binPath, args, parser.FeedLine)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Handle browser cookies database lock
		if parser.hasCookieLock {
			return nil, fmt.Errorf("Trình duyệt đang mở và khóa file cookies. Cậu chủ vui lòng đóng trình duyệt rồi thử lại.")
		}

		if opts.DownloadPlaylist && len(parser.downloadedFiles) > 0 {
			// Playlist completed with some videos downloaded; ignore non-zero exit
		} else {
			return nil, fmt.Errorf("lỗi trong quá trình tải xuống: %w", err)
		}
	}

	return parser.Result()
}
