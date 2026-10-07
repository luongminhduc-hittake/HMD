package downloader

import (
	"context"
	"path/filepath"
	"testing"

	"hmd/internal/spotify"
)

func TestInjectMetadata_NoFFmpegOrFile(t *testing.T) {
	// If ffmpegPath is empty, should cleanly return nil
	err := InjectMetadata(context.Background(), "", "/dummy/path.mp3", spotify.TrackInfo{Title: "Test"})
	if err != nil {
		t.Errorf("expected nil error when ffmpegPath is empty, got: %v", err)
	}

	// If audio file doesn't exist
	err = InjectMetadata(context.Background(), "/usr/bin/ffmpeg", "/nonexistent/audio.mp3", spotify.TrackInfo{Title: "Test"})
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestBuildDownloadArgs_CustomFilename(t *testing.T) {
	opts := DownloadOptions{
		OutputDir:      "/tmp/music",
		CustomFilename: "Artist - Song With: Special? Chars",
		Preset:         PresetAudioMP3,
		URL:            "ytsearch1:Artist Song",
	}

	args := BuildDownloadArgs(opts)
	argStr := ""
	for _, a := range args {
		argStr += a + " "
	}

	expectedPattern := filepath.Join("/tmp/music", "Artist - Song With_ Special_ Chars.%(ext)s")
	if !filepath.IsAbs(expectedPattern) {
		t.Errorf("expected absolute path")
	}
}
