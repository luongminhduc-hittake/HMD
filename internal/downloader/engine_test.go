package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestBuildDownloadArgs(t *testing.T) {
	opts := DownloadOptions{
		URL:              "https://youtube.com/watch?v=12345",
		OutputDir:        "/tmp/test_download",
		Preset:           PresetBestVideo,
		DownloadPlaylist: false,
		TrimRange:        "01:00-02:30",
		EnableSubtitles:  true,
		FFmpegPath:       "/usr/bin/ffmpeg",
	}

	args := BuildDownloadArgs(opts)
	argStr := strings.Join(args, " ")

	if !strings.Contains(argStr, "--download-sections *01:00-02:30") {
		t.Errorf("expected trim range flag, got args: %s", argStr)
	}
	if !strings.Contains(argStr, "--force-keyframes-at-cuts") {
		t.Errorf("expected --force-keyframes-at-cuts, got args: %s", argStr)
	}
	if !strings.Contains(argStr, "--write-subs") || !strings.Contains(argStr, "--embed-subs") {
		t.Errorf("expected subtitle flags, got args: %s", argStr)
	}
	if !strings.Contains(argStr, "--ffmpeg-location /usr/bin") {
		t.Errorf("expected ffmpeg location, got args: %s", argStr)
	}
	if !strings.Contains(argStr, "--no-playlist") {
		t.Errorf("expected --no-playlist, got args: %s", argStr)
	}
	if !strings.HasSuffix(argStr, "-- https://youtube.com/watch?v=12345") {
		t.Errorf("expected '-- URL' to be at the end, got args: %s", argStr)
	}
}

func TestBuildDownloadArgsPresets(t *testing.T) {
	audioOpts := DownloadOptions{
		Preset: PresetAudioMP3,
	}
	audioArgs := strings.Join(BuildDownloadArgs(audioOpts), " ")
	if !strings.Contains(audioArgs, "--audio-format mp3") {
		t.Errorf("expected mp3 format, got %s", audioArgs)
	}

	playlistOpts := DownloadOptions{
		DownloadPlaylist: true,
	}
	playlistArgs := strings.Join(BuildDownloadArgs(playlistOpts), " ")
	if !strings.Contains(playlistArgs, "--yes-playlist") {
		t.Errorf("expected --yes-playlist, got %s", playlistArgs)
	}

	// Test WAV preset: ensure no --embed-thumbnail
	wavOpts := DownloadOptions{
		Preset: PresetAudioWAV,
	}
	wavArgs := strings.Join(BuildDownloadArgs(wavOpts), " ")
	if !strings.Contains(wavArgs, "--audio-format wav") {
		t.Errorf("expected wav format, got %s", wavArgs)
	}
	if strings.Contains(wavArgs, "--embed-thumbnail") {
		t.Errorf("WAV preset must NOT contain --embed-thumbnail, got %s", wavArgs)
	}

	// Test Thumbnail preset
	thumbOpts := DownloadOptions{
		Preset: PresetThumbnail,
	}
	thumbArgs := strings.Join(BuildDownloadArgs(thumbOpts), " ")
	if !strings.Contains(thumbArgs, "--write-thumbnail") || !strings.Contains(thumbArgs, "--skip-download") {
		t.Errorf("expected thumbnail flags, got %s", thumbArgs)
	}

	// Test Cookies & Security flags
	secOpts := DownloadOptions{
		CookiesBrowser:      "chrome",
		ConcurrentFragments: 6,
		MaxRetries:          5,
	}
	secArgs := strings.Join(BuildDownloadArgs(secOpts), " ")
	if !strings.Contains(secArgs, "--no-config") {
		t.Errorf("expected --no-config, got %s", secArgs)
	}
	if !strings.Contains(secArgs, "--windows-filenames") {
		t.Errorf("expected --windows-filenames, got %s", secArgs)
	}
	if !strings.Contains(secArgs, "--cookies-from-browser chrome") {
		t.Errorf("expected --cookies-from-browser chrome, got %s", secArgs)
	}
	if !strings.Contains(secArgs, "--concurrent-fragments 6") {
		t.Errorf("expected --concurrent-fragments 6, got %s", secArgs)
	}

	// Test Unsafe Browser rejected
	unsafeOpts := DownloadOptions{
		CookiesBrowser: "chrome; rm -rf /",
	}
	unsafeArgs := strings.Join(BuildDownloadArgs(unsafeOpts), " ")
	if strings.Contains(unsafeArgs, "--cookies-from-browser") {
		t.Errorf("unsafe browser should not be accepted, got %s", unsafeArgs)
	}
}

func fakeExecCommandContext(mode string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", mode}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "HELPER_MODE=" + mode}
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	mode := os.Getenv("HELPER_MODE")
	switch mode {
	case "success":
		fmt.Println("DOWNLOAD_PROGRESS: 50.0%|1.5MiB/s|00:10|10.5MiB")
		fmt.Println("DOWNLOAD_PROGRESS: 100.0%|2.0MiB/s|00:00|21.0MiB")
		fmt.Println("FINAL_PATH:/tmp/test_download/video.mp4")
		os.Exit(0)
	case "cookie_locked":
		fmt.Fprintln(os.Stderr, "ERROR: [youtube] Could not copy Chrome cookie database (database is locked)")
		os.Exit(1)
	case "thumbnail":
		fmt.Println("[info] Writing video thumbnail 1 to: /tmp/test_download/cover.jpg")
		os.Exit(0)
	case "hang":
		time.Sleep(10 * time.Second)
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

func TestExecuteDownloadMockSuccess(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()
	execCommandContext = fakeExecCommandContext("success")

	var lastProgress ProgressUpdate
	res, err := ExecuteDownload(context.Background(), DownloadOptions{
		OutputDir: t.TempDir(),
		Preset:    Preset1080p,
		OnProgress: func(pu ProgressUpdate) {
			lastProgress = pu
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FilePath != "/tmp/test_download/video.mp4" {
		t.Errorf("expected final path /tmp/test_download/video.mp4, got %s", res.FilePath)
	}
	if lastProgress.Percent != 100.0 {
		t.Errorf("expected last percent 100.0, got %f", lastProgress.Percent)
	}
}

func TestExecuteDownloadMockCookieLocked(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()
	execCommandContext = fakeExecCommandContext("cookie_locked")

	_, err := ExecuteDownload(context.Background(), DownloadOptions{
		OutputDir:      t.TempDir(),
		CookiesBrowser: "chrome",
	})

	if err == nil {
		t.Fatalf("expected error on cookie lock, got nil")
	}
	if !strings.Contains(err.Error(), "Trình duyệt đang mở và khóa file cookies") {
		t.Errorf("expected friendly cookie locked message, got: %v", err)
	}
}

func TestExecuteDownloadMockThumbnail(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()
	execCommandContext = fakeExecCommandContext("thumbnail")

	var lastProgress ProgressUpdate
	res, err := ExecuteDownload(context.Background(), DownloadOptions{
		OutputDir: t.TempDir(),
		Preset:    PresetThumbnail,
		OnProgress: func(pu ProgressUpdate) {
			lastProgress = pu
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FilePath != "/tmp/test_download/cover.jpg" {
		t.Errorf("expected thumbnail path /tmp/test_download/cover.jpg, got %s", res.FilePath)
	}
	if res.DownloadedCount != 1 {
		t.Errorf("expected downloaded count 1, got %d", res.DownloadedCount)
	}
	if lastProgress.Percent != 100.0 {
		t.Errorf("expected progress 100.0, got %f", lastProgress.Percent)
	}
}

func TestExecuteDownloadMockCancel(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()
	execCommandContext = fakeExecCommandContext("hang")

	ctx, cancel := context.WithCancel(context.Background())
	// cancel shortly after starting
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := ExecuteDownload(ctx, DownloadOptions{
		OutputDir: t.TempDir(),
	})

	if err == nil {
		t.Fatalf("expected error on canceled context, got nil")
	}
}

func TestResolveCookieArg(t *testing.T) {
	if got := resolveCookieArg(""); got != "" {
		t.Errorf("expected empty string for empty input, got %q", got)
	}
	if got := resolveCookieArg("firefox:/custom/profile"); got != "firefox:/custom/profile" {
		t.Errorf("expected custom firefox profile unchanged, got %q", got)
	}
	if got := resolveCookieArg("chromium:/custom/profile"); got != "chromium:/custom/profile" {
		t.Errorf("expected custom chromium profile unchanged, got %q", got)
	}
	// Chrome should resolve to something valid (either detected or "chrome")
	if got := resolveCookieArg("chrome"); got == "" {
		t.Errorf("expected non-empty resolved chrome cookie arg")
	}
}
