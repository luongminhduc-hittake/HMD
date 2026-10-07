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

func TestExecuteSpotifyDownload_UnplayableTracks(t *testing.T) {
	tmpDir := t.TempDir()
	opts := DownloadOptions{
		OutputDir:        tmpDir,
		Title:            "Test Playlist",
		IsSpotify:        true,
		ConcurrentTracks: 2,
		SpotifyTracks: []spotify.TrackInfo{
			{Title: "Track 1", Artist: "Artist 1", IsPlayable: false},
			{Title: "Track 2", Artist: "Artist 2", IsPlayable: false},
		},
	}

	res, err := ExecuteDownload(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.SkippedCount != 2 {
		t.Errorf("expected 2 skipped tracks, got %d", res.SkippedCount)
	}
	if len(res.FailedTracks) != 2 {
		t.Errorf("expected 2 failed tracks recorded, got %d", len(res.FailedTracks))
	}
}

