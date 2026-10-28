package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYtDlpParser_ProgressParsing(t *testing.T) {
	var lastUpdate ProgressUpdate
	opts := DownloadOptions{
		Title: "Test Video",
		OnProgress: func(pu ProgressUpdate) {
			lastUpdate = pu
		},
	}

	parser := NewYtDlpParser(opts)

	// Feed progress lines
	parser.FeedLine("DOWNLOAD_PROGRESS: 45.2%|1.2MiB/s|00:15|10.5MiB")

	if lastUpdate.Percent != 45.2 {
		t.Errorf("expected percent 45.2, got %f", lastUpdate.Percent)
	}
	if lastUpdate.Speed != "1.2MiB/s" {
		t.Errorf("expected speed '1.2MiB/s', got %q", lastUpdate.Speed)
	}
	if lastUpdate.ETA != "00:15" {
		t.Errorf("expected ETA '00:15', got %q", lastUpdate.ETA)
	}
	if lastUpdate.TotalSize != "10.5MiB" {
		t.Errorf("expected TotalSize '10.5MiB', got %q", lastUpdate.TotalSize)
	}

	// Feed NA values
	parser.FeedLine("DOWNLOAD_PROGRESS: 100.0%|NA|NA|NA")
	if lastUpdate.Percent != 100.0 {
		t.Errorf("expected percent 100.0, got %f", lastUpdate.Percent)
	}
	if lastUpdate.Speed != "" {
		t.Errorf("expected empty speed for NA, got %q", lastUpdate.Speed)
	}
}

func TestYtDlpParser_PathDetection(t *testing.T) {
	opts := DownloadOptions{
		Title: "Test Video",
	}
	parser := NewYtDlpParser(opts)

	parser.FeedLine("FINAL_PATH:/downloads/video.mp4")
	if parser.finalFilePath != "/downloads/video.mp4" {
		t.Errorf("expected path '/downloads/video.mp4', got %q", parser.finalFilePath)
	}

	res, err := parser.Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FileName != "video.mp4" {
		t.Errorf("expected filename 'video.mp4', got %q", res.FileName)
	}
	if res.DownloadedCount != 1 {
		t.Errorf("expected downloaded count 1, got %d", res.DownloadedCount)
	}
}

func TestYtDlpParser_PostProcessingAndCookieLock(t *testing.T) {
	var statusMsg string
	opts := DownloadOptions{
		Title: "Test Video",
		OnStatus: func(s string) {
			statusMsg = s
		},
	}
	parser := NewYtDlpParser(opts)

	parser.FeedLine("[Merger] Merging formats into \"output.mp4\"")
	if statusMsg != "Đang hợp nhất video và audio..." {
		t.Errorf("expected merger status message, got %q", statusMsg)
	}

	parser.FeedLine("ERROR: database is locked")
	if !parser.hasCookieLock {
		t.Errorf("expected hasCookieLock to be true")
	}
}

func TestGalleryParser(t *testing.T) {
	var statusMsg string
	var lastProg ProgressUpdate
	opts := DownloadOptions{
		Title:     "Photo Album",
		OutputDir: "/downloads",
		OnStatus: func(s string) {
			statusMsg = s
		},
		OnProgress: func(pu ProgressUpdate) {
			lastProg = pu
		},
	}
	parser := NewGalleryParser(opts)

	parser.FeedLine("/downloads/image1.jpg")
	parser.FeedLine("/downloads/image2.jpg")

	if parser.finalFilePath != "/downloads/image2.jpg" {
		t.Errorf("expected final path image2.jpg, got %q", parser.finalFilePath)
	}
	if statusMsg != "Đã tải: image2.jpg" {
		t.Errorf("expected status 'Đã tải: image2.jpg', got %q", statusMsg)
	}

	res, err := parser.Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DownloadedCount != 2 {
		t.Errorf("expected downloaded count 2, got %d", res.DownloadedCount)
	}
	if lastProg.Percent != 100 {
		t.Errorf("expected percent 100, got %f", lastProg.Percent)
	}
}

func TestGalleryParserSkipsAlreadyDownloadedPrefix(t *testing.T) {
	var statuses []string
	opts := DownloadOptions{
		Title:     "Photo Album",
		OutputDir: "/downloads",
		OnStatus:  func(s string) { statuses = append(statuses, s) },
	}
	parser := NewGalleryParser(opts)

	// gallery-dl prints "# <path>" when the file already exists on disk.
	parser.FeedLine("# /downloads/127.0.0.1__pic.jpg")

	if parser.finalFilePath != "/downloads/127.0.0.1__pic.jpg" {
		t.Errorf("expected final path from '#' line, got %q", parser.finalFilePath)
	}

	res, err := parser.Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DownloadedCount != 1 {
		t.Errorf("expected downloaded count 1, got %d", res.DownloadedCount)
	}
}

func TestGalleryParserEmptyResultIsError(t *testing.T) {
	parser := NewGalleryParser(DownloadOptions{Title: "Empty Album", OutputDir: "/downloads"})

	res, err := parser.Result()
	if err == nil {
		t.Fatalf("expected error when gallery-dl reported no files, got result %+v", res)
	}
}

func TestYtDlpParserDownloadedCountCountsItemsNotStreams(t *testing.T) {
	parser := NewYtDlpParser(DownloadOptions{DownloadPlaylist: true})

	// Two playlist items, each with a video + audio stream that gets merged.
	lines := []string{
		"[download] Downloading item 1 of 2",
		"[download] Destination: /dl/Video One.f137.mp4",
		"[download] Destination: /dl/Video One.f251.webm",
		"[Merger] Merging formats into \"/dl/Video One.mp4\"",
		"FINAL_PATH:/dl/Video One.mp4",
		"[download] Downloading item 2 of 2",
		"[download] Destination: /dl/Video Two.f137.mp4",
		"[download] Destination: /dl/Video Two.f251.webm",
		"[Merger] Merging formats into \"/dl/Video Two.mp4\"",
		"FINAL_PATH:/dl/Video Two.mp4",
	}
	for _, l := range lines {
		parser.FeedLine(l)
	}

	res, err := parser.Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DownloadedCount != 2 {
		t.Errorf("expected 2 downloaded items for a 2-item playlist, got %d", res.DownloadedCount)
	}
}

func TestYtDlpParserThumbnailUsesConvertedJPGExtension(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "cover.webp")
	jpg := filepath.Join(dir, "cover.jpg")
	if err := os.WriteFile(jpg, []byte("jpeg-bytes"), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYtDlpParser(DownloadOptions{Preset: PresetThumbnail, OutputDir: dir})
	// Real yt-dlp sequence: writes .webp, converts to .jpg, deletes the .webp.
	parser.FeedLine("[info] Writing video thumbnail 41 to: " + src)
	parser.FeedLine("[ThumbnailsConvertor] Converting thumbnail \"" + src + "\" to jpg")
	parser.FeedLine("Deleting original file " + src + " (pass -k to keep)")

	res, err := parser.Result()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FilePath != jpg {
		t.Errorf("expected converted thumbnail path %q, got %q", jpg, res.FilePath)
	}
	if res.FileSize != int64(len("jpeg-bytes")) {
		t.Errorf("expected file size of the .jpg, got %d", res.FileSize)
	}
}
