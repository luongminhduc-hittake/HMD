package downloader

import (
	"context"
	"fmt"

	"hmd/internal/util"
)

// runStandardDownload handles yt-dlp execution for single videos and playlists,
// with multi-browser cookie cascade fallback when cookie issues or locks occur.
func runStandardDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	candidates := util.GetCookieCascade(opts.CookiesBrowser)
	if len(candidates) == 0 {
		return runSingleStandardDownload(ctx, opts)
	}

	for i, cand := range candidates {
		tryOpts := opts
		tryOpts.CookiesBrowser = cand.CookieArg

		res, parser, err := executeStandardAttempt(ctx, tryOpts)
		if err == nil {
			return res, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		isRetryable := parser.hasCookieLock || parser.hasAuthError
		if isRetryable && i+1 < len(candidates) {
			if opts.OnStatus != nil {
				opts.OnStatus(fmt.Sprintf("Cookies từ %s gặp lỗi/khóa, đang chuyển sang %s...", cand.Name, candidates[i+1].Name))
			}
			continue
		}

		if parser.hasCookieLock {
			return nil, fmt.Errorf("Trình duyệt đang mở và khóa file cookies. Cậu chủ vui lòng đóng trình duyệt rồi thử lại.")
		}

		if tryOpts.DownloadPlaylist && len(parser.downloadedFiles) > 0 {
			return parser.Result()
		}

		return nil, fmt.Errorf("lỗi trong quá trình tải xuống: %w", err)
	}

	return nil, fmt.Errorf("không có cấu hình cookie khả dụng")
}

func executeStandardAttempt(ctx context.Context, opts DownloadOptions) (*DownloadResult, *YtDlpParser, error) {
	binPath := opts.YtDlpPath
	if binPath == "" {
		binPath = "yt-dlp"
	}

	args := BuildDownloadArgs(opts)
	parser := NewYtDlpParser(opts)

	err := streamCommand(ctx, binPath, args, parser.FeedLine)
	if err != nil {
		return nil, parser, err
	}
	res, err := parser.Result()
	return res, parser, err
}

func runSingleStandardDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	res, parser, err := executeStandardAttempt(ctx, opts)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if parser.hasCookieLock {
			return nil, fmt.Errorf("Trình duyệt đang mở và khóa file cookies. Cậu chủ vui lòng đóng trình duyệt rồi thử lại.")
		}
		if opts.DownloadPlaylist && len(parser.downloadedFiles) > 0 {
			return parser.Result()
		}
		return nil, fmt.Errorf("lỗi trong quá trình tải xuống: %w", err)
	}
	return res, nil
}
