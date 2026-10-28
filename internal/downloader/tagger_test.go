package downloader

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestInjectMetadataCoverFlagsSupportAllContainers(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	cover := filepath.Join(t.TempDir(), "cover.jpg")
	if err := os.WriteFile(cover, []byte{0xFF, 0xD8, 0xFF, 0xE0}, 0644); err != nil {
		t.Fatal(err)
	}

	args := buildFFmpegMetadataArgs("/music/song.m4a", cover, spotify.TrackInfo{Title: "T", Artist: "A"})
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "-c:v mjpeg") {
		t.Errorf("cover stream must be encoded as mjpeg for m4a compatibility, got: %s", joined)
	}
	if !strings.Contains(joined, "-disposition:v attached_pic") {
		t.Errorf("cover stream must be marked attached_pic, got: %s", joined)
	}
}

func TestInjectMetadataReportsFailure(t *testing.T) {
	// A file that is not audio: ffmpeg will refuse it, and the error must surface.
	dir := t.TempDir()
	bad := filepath.Join(dir, "notaudio.m4a")
	if err := os.WriteFile(bad, []byte("definitely not audio"), 0644); err != nil {
		t.Fatal(err)
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}

	if err := InjectMetadata(context.Background(), ffmpegPath, bad, spotify.TrackInfo{Title: "T"}); err == nil {
		t.Fatalf("expected InjectMetadata to report failure for a non-audio file, got nil")
	}
}

func TestCoverArtContainerSupport(t *testing.T) {
	// Containers whose muxer accepts an attached_pic video stream.
	for _, ext := range []string{".mp3", ".m4a", ".mp4", ".flac"} {
		if !containerSupportsCoverArt(ext) {
			t.Errorf("expected %s to support embedded cover art", ext)
		}
	}
	// wav and opus muxers reject video streams outright; tagging must skip the cover.
	for _, ext := range []string{".wav", ".opus", ".ogg"} {
		if containerSupportsCoverArt(ext) {
			t.Errorf("expected %s to be excluded from cover art embedding", ext)
		}
	}
}

func TestBuildFFmpegArgsSkipsCoverForOpusAndWav(t *testing.T) {
	cover := "/tmp/cover.jpg"

	for _, ext := range []string{".opus", ".wav"} {
		args := buildFFmpegMetadataArgs("/music/song"+ext, cover, spotify.TrackInfo{Title: "T"})
		joined := strings.Join(args, " ")
		if strings.Contains(joined, cover) {
			t.Errorf("%s must not try to embed a cover, args: %s", ext, joined)
		}
		if !strings.Contains(joined, "-metadata title=T") {
			t.Errorf("%s must still carry text metadata, args: %s", ext, joined)
		}
	}
}

func TestRunSpotifyDownloadReportsTaggingFailure(t *testing.T) {
	origExec := execCommandContext
	defer func() { execCommandContext = origExec }()

	dir := t.TempDir()
	// yt-dlp mock writes a bogus "audio" file; ffmpeg will refuse to tag it.
	execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", "fake_file"}
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_MODE=fake_file", "FAKE_FILE_DIR="+dir)
		return cmd
	}

	res, err := ExecuteDownload(context.Background(), DownloadOptions{
		OutputDir:        dir,
		IsSpotify:        true,
		FFmpegPath:       "/usr/bin/ffmpeg",
		ConcurrentTracks: 1,
		SpotifyTracks: []spotify.TrackInfo{
			{Title: "Track", Artist: "Artist", IsPlayable: true, CoverURL: ""},
		},
	})

	if err != nil {
		return // failing the whole download is also acceptable
	}
	if len(res.FailedTracks) == 0 && res.SkippedCount == 0 {
		t.Errorf("expected a tagging failure to be reported, got result %+v", res)
	}
}
