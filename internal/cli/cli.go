package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"

	"ytdownloader/internal/deps"
	"ytdownloader/internal/downloader"
	"ytdownloader/internal/util"
)

const AppVersion = "1.0.0"

const BannerASCII = `
    __    _ __  __        __               __  __ ______ ____  __ 
   / /_  (_) /_/ /_____ _/ /_____   '__   / / / //_  __// __ \/ / 
  / __ \/ / __/ __/ __ '/ //_/ _ \  / _\ / /_/ /  / /  / / / / /  
 / / / / / /_/ /_/ /_/ / ,< /  __/ /__ \ \__, /  / /  / /_/ / /___
/_/ /_/_/\__/\__/\__,_/_/|_|\___/  \___//____/  /_/  /_____/_____/`

func renderLogo() string {
	lines := strings.Split(strings.Trim(BannerASCII, "\n"), "\n")
	colors := []string{
		"#00E5FF",
		"#00C6FF",
		"#3B82F6",
		"#8B5CF6",
		"#A855F7",
	}
	var b strings.Builder
	for i, line := range lines {
		c := colors[i%len(colors)]
		st := lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Bold(true)
		b.WriteString(st.Render(line) + "\n")
	}
	return b.String()
}

func renderProgressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	unfilled := width - filled

	fillStr := strings.Repeat("█", filled)
	unfillStr := strings.Repeat("░", unfilled)

	var coloredFill strings.Builder
	runes := []rune(fillStr)
	total := len(runes)
	for i, r := range runes {
		ratio := float64(i) / float64(total)
		var c string
		if ratio < 0.35 {
			c = "#00E5FF"
		} else if ratio < 0.70 {
			c = "#3B82F6"
		} else {
			c = "#A855F7"
		}
		coloredFill.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(string(r)))
	}

	dimUnfill := lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563")).Render(unfillStr)
	return "[" + coloredFill.String() + dimUnfill + "]"
}

// Run handles execution when CLI arguments/flags are provided.
func Run(args []string) int {
	fs := flag.NewFlagSet("ytdl", flag.ExitOnError)

	var (
		urlFlag         = fs.String("u", "", "URL video/audio YouTube")
		urlLongFlag     = fs.String("url", "", "URL video/audio YouTube (viết dài)")
		formatFlag      = fs.String("f", "best", "Định dạng tải về (best, 1080p, 720p, mp3, m4a)")
		formatLongFlag  = fs.String("format", "", "Định dạng tải về (viết dài)")
		outputFlag      = fs.String("o", "", "Thư mục lưu tệp")
		outputLongFlag  = fs.String("output", "", "Thư mục lưu tệp (viết dài)")
		playlistFlag    = fs.Bool("playlist", false, "Tải toàn bộ playlist nếu có")
		noPlaylistFlag  = fs.Bool("no-playlist", true, "Chỉ tải video đơn lẻ")
		openFlag        = fs.Bool("open", false, "Mở thư mục chứa tệp sau khi tải xong")
		updateFlag      = fs.Bool("update", false, "Cập nhật yt-dlp lên bản mới nhất")
		versionFlag     = fs.Bool("v", false, "Xem phiên bản")
		versionLongFlag = fs.Bool("version", false, "Xem phiên bản (viết dài)")
	)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi phân tích cú pháp cờ lệnh: %v\n", err)
		return 1
	}

	if *versionFlag || *versionLongFlag {
		fmt.Printf("hittake's ytdl v%s (%s/%s)\n", AppVersion, runtime.GOOS, runtime.GOARCH)
		return 0
	}

	// 1. Dependency check (Silent if already satisfied)
	paths, missing := deps.FindBinaries()
	if len(missing) > 0 {
		fmt.Println("⚡ Đang chuẩn bị công cụ hệ thống...")
		var err error
		paths, err = deps.EnsureDependencies(func(name string, dl, total int64, pct float64) {
			if total > 0 {
				fmt.Printf("\r⬇ Đang tải %s: %.1f%% (%d/%d MB)", name, pct, dl/(1024*1024), total/(1024*1024))
			}
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n✖ Lỗi chuẩn bị công cụ: %v\n", err)
			return 1
		}
		fmt.Println("\r✔ Các công cụ (yt-dlp, ffmpeg) đã sẵn sàng!           ")
	}

	if *updateFlag {
		fmt.Print(renderLogo())
		fmt.Println("🔄 Đang cập nhật yt-dlp...")
		out, err := deps.UpdateYtDlp()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✖ Cập nhật thất bại: %v\n", err)
			return 1
		}
		fmt.Println(out)
		return 0
	}

	// Determine URL from -u, --url or positional argument
	rawURL := *urlFlag
	if *urlLongFlag != "" {
		rawURL = *urlLongFlag
	}
	if rawURL == "" {
		urlArgs := fs.Args()
		if len(urlArgs) > 0 {
			rawURL = urlArgs[0]
		}
	}

	if rawURL == "" {
		fmt.Print(renderLogo())
		fmt.Println()
		fmt.Fprintln(os.Stderr, "✖ Vui lòng cung cấp link video/audio YouTube.")
		fmt.Println("Ví dụ: ytdl \"https://www.youtube.com/watch?v=...\"")
		fmt.Println("       ytdl -f mp3 \"https://www.youtube.com/watch?v=...\"")
		fmt.Println("       ytdl -u \"https://www.youtube.com/watch?v=...\" -f 1080p -o ~/Videos")
		return 1
	}

	// Format preset resolution
	formatChoice := *formatFlag
	if *formatLongFlag != "" {
		formatChoice = *formatLongFlag
	}

	var presetName string
	preset := downloader.PresetBestVideo
	switch strings.ToLower(formatChoice) {
	case "best", "mp4":
		preset = downloader.PresetBestVideo
		presetName = "Best Video (MP4)"
	case "1080", "1080p":
		preset = downloader.Preset1080p
		presetName = "Full HD 1080p (MP4)"
	case "720", "720p":
		preset = downloader.Preset720p
		presetName = "HD 720p (MP4)"
	case "mp3", "audio":
		preset = downloader.PresetAudioMP3
		presetName = "Audio MP3 (320kbps)"
	case "m4a", "aac":
		preset = downloader.PresetAudioM4A
		presetName = "Audio M4A (AAC)"
	default:
		preset = downloader.PresetBestVideo
		presetName = "Best Video (MP4)"
	}

	// Output dir resolution
	destDir := *outputFlag
	if *outputLongFlag != "" {
		destDir = *outputLongFlag
	}
	if destDir == "" {
		destDir = util.GetDefaultDownloadDir()
	}

	downloadPlaylist := *playlistFlag
	if !*playlistFlag && *noPlaylistFlag {
		downloadPlaylist = false
	}

	// Print Logo
	fmt.Print(renderLogo())
	fmt.Println()

	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E5FF")).Bold(true)
	fmt.Printf("🔍 Đang tải: %s\n\n", urlStyle.Render(rawURL))

	// In-place dynamic 2-line dashboard state
	var mu sync.Mutex
	hasDrawn := false
	currentStage := "Đang kết nối..."
	lastSpeed := "---"
	lastETA := "---"
	lastTotal := "---"

	drawDashboard := func(pu downloader.ProgressUpdate) {
		mu.Lock()
		defer mu.Unlock()

		pct := pu.Percent
		if pct > 100 {
			pct = 100
		}

		bar := renderProgressBar(pct, 26)
		pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E5FF")).Bold(true)
		stageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))

		stageText := currentStage
		if pu.StatusMessage != "" && currentStage == "Đang kết nối..." {
			stageText = pu.StatusMessage
		}

		line1 := fmt.Sprintf("%s %s  %s", bar, pctStyle.Render(fmt.Sprintf("%5.1f%%", pct)), stageStyle.Render(stageText))

		if pu.Speed != "" {
			lastSpeed = pu.Speed
		}
		if pu.ETA != "" {
			lastETA = pu.ETA
		}
		if pu.TotalSize != "" {
			lastTotal = pu.TotalSize
		}

		badgeSpeed := lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")).Render(fmt.Sprintf("⚡ %-10s", lastSpeed))
		badgeETA := lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")).Render(fmt.Sprintf("⏳ Còn lại: %-8s", lastETA))
		badgeSize := lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Render(fmt.Sprintf("📦 %s", lastTotal))

		line2 := fmt.Sprintf("%s   %s   %s", badgeSpeed, badgeETA, badgeSize)

		if hasDrawn {
			fmt.Printf("\033[1A\033[2K\r%s\n\033[2K\r%s", line1, line2)
		} else {
			fmt.Printf("%s\n%s", line1, line2)
			hasDrawn = true
		}
	}

	opts := downloader.DownloadOptions{
		YtDlpPath:        paths.YtDlp,
		FFmpegPath:       paths.FFmpeg,
		URL:              rawURL,
		OutputDir:        destDir,
		Preset:           preset,
		DownloadPlaylist: downloadPlaylist,
		OnProgress: func(pu downloader.ProgressUpdate) {
			drawDashboard(pu)
		},
		OnStatus: func(st string) {
			mu.Lock()
			currentStage = st
			mu.Unlock()
			drawDashboard(downloader.ProgressUpdate{Percent: 100.0, StatusMessage: st})
		},
	}

	res, err := downloader.ExecuteDownload(context.Background(), opts)
	if err != nil {
		if hasDrawn {
			fmt.Printf("\033[1A\033[2K\r\033[1B\033[2K\r\033[1A")
		}
		errStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BF616A")).
			Padding(1, 2)
		fmt.Println(errStyle.Render(fmt.Sprintf("✖ LỖI KHI TẢI XUỐNG:\n%v", err)))
		return 1
	}

	// Clear the 2-line dashboard
	if hasDrawn {
		fmt.Printf("\033[1A\033[2K\r\033[1B\033[2K\r\033[1A")
	}

	// Success Card
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00E5FF")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#10B981"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#81A1C1")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("✔ TẢI XUỐNG THÀNH CÔNG") + "\n\n")

	if res.FileName != "" {
		sizeStr := ""
		if res.FormattedSize != "" {
			sizeStr = fmt.Sprintf(" (%s)", res.FormattedSize)
		}
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("📦 Tệp tin:  "), valueStyle.Render(res.FileName+sizeStr)))
	}
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("🎯 Định dạng:"), valueStyle.Render(presetName)))
	if res.FilePath != "" {
		content.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("📂 Thư mục:  "), valueStyle.Render(res.FilePath)))
	}

	fmt.Println(cardStyle.Render(content.String()))

	if *openFlag {
		target := res.FilePath
		if target == "" {
			target = destDir
		}
		_ = util.OpenFolder(target)
	}

	return 0
}
