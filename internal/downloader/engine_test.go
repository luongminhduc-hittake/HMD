package downloader

import (
	"strings"
	"testing"
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
}
