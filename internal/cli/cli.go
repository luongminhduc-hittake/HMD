package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"ytdownloader/internal/deps"
	"ytdownloader/internal/downloader"
)

const AppVersion = "1.0.0"

// Run handles execution when CLI arguments/flags are provided.
func Run(args []string) int {
	fs := flag.NewFlagSet("ytdl", flag.ExitOnError)

	var (
		formatFlag      = fs.String("f", "best", "Định dạng tải về (best, 1080p, 720p, mp3, m4a)")
		formatLongFlag  = fs.String("format", "", "Định dạng tải về (viết dài)")
		outputFlag      = fs.String("o", ".", "Thư mục lưu tệp")
		outputLongFlag  = fs.String("output", "", "Thư mục lưu tệp (viết dài)")
		playlistFlag    = fs.Bool("playlist", false, "Tải toàn bộ playlist nếu có")
		noPlaylistFlag  = fs.Bool("no-playlist", true, "Chỉ tải video đơn lẻ")
		updateFlag      = fs.Bool("update", false, "Cập nhật yt-dlp lên bản mới nhất")
		versionFlag     = fs.Bool("v", false, "Xem phiên bản")
		versionLongFlag = fs.Bool("version", false, "Xem phiên bản (viết dài)")
	)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi phân tích cú pháp cờ lệnh: %v\n", err)
		return 1
	}

	if *versionFlag || *versionLongFlag {
		fmt.Printf("ytdl v%s (Windows Edition)\n", AppVersion)
		return 0
	}

	// 1. Dependency check
	fmt.Println("⚡ Đang kiểm tra công cụ hệ thống...")
	paths, err := deps.EnsureDependencies(func(name string, dl, total int64, pct float64) {
		if total > 0 {
			fmt.Printf("\r⬇ Đang tải %s: %.1f%% (%d/%d MB)", name, pct, dl/(1024*1024), total/(1024*1024))
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✖ Lỗi chuẩn bị công cụ: %v\n", err)
		return 1
	}
	fmt.Println("\r✔ Các công cụ (yt-dlp, ffmpeg) đã sẵn sàng!           ")

	if *updateFlag {
		fmt.Println("🔄 Đang cập nhật yt-dlp...")
		out, err := deps.UpdateYtDlp()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✖ Cập nhật thất bại: %v\n", err)
			return 1
		}
		fmt.Println(out)
		return 0
	}

	// Determine URL
	urlArgs := fs.Args()
	if len(urlArgs) == 0 {
		fmt.Fprintln(os.Stderr, "✖ Vui lòng cung cấp link video/audio YouTube.")
		fmt.Println("Ví dụ: ytdl -f mp3 \"https://www.youtube.com/watch?v=...\"")
		return 1
	}
	rawURL := urlArgs[0]

	// Format preset resolution
	formatChoice := *formatFlag
	if *formatLongFlag != "" {
		formatChoice = *formatLongFlag
	}

	preset := downloader.PresetBestVideo
	switch strings.ToLower(formatChoice) {
	case "best", "mp4":
		preset = downloader.PresetBestVideo
	case "1080", "1080p":
		preset = downloader.Preset1080p
	case "720", "720p":
		preset = downloader.Preset720p
	case "mp3", "audio":
		preset = downloader.PresetAudioMP3
	case "m4a", "aac":
		preset = downloader.PresetAudioM4A
	default:
		fmt.Printf("⚠ Định dạng '%s' không rõ, tự động chuyển về 'best (MP4)'\n", formatChoice)
	}

	// Output dir resolution
	destDir := *outputFlag
	if *outputLongFlag != "" {
		destDir = *outputLongFlag
	}

	downloadPlaylist := *playlistFlag
	if !*playlistFlag && *noPlaylistFlag {
		downloadPlaylist = false
	}

	fmt.Printf("🔍 Đang tải liên kết: %s\n", rawURL)
	opts := downloader.DownloadOptions{
		YtDlpPath:        paths.YtDlp,
		FFmpegPath:       paths.FFmpeg,
		URL:              rawURL,
		OutputDir:        destDir,
		Preset:           preset,
		DownloadPlaylist: downloadPlaylist,
		OnProgress: func(pu downloader.ProgressUpdate) {
			fmt.Printf("\r[ytdl] %.1f%% │ Tốc độ: %-10s │ ETA: %-8s │ Dung lượng: %s    ",
				pu.Percent, pu.Speed, pu.ETA, pu.TotalSize)
		},
		OnStatus: func(st string) {
			fmt.Printf("\n⚙ %s\n", st)
		},
	}

	res, err := downloader.ExecuteDownload(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✖ Lỗi khi tải video: %v\n", err)
		return 1
	}

	fmt.Println("\n\n✔ TẢI XUỐNG THÀNH CÔNG!")
	if res.FileName != "" {
		fmt.Printf("📦 Tệp tin:   %s (%s)\n", res.FileName, res.FormattedSize)
	}
	if res.FilePath != "" {
		fmt.Printf("📂 Đường dẫn: %s\n", res.FilePath)
	}
	return 0
}
