package downloader

import (
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
