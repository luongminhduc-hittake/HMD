package deps

import (
	"strings"
	"testing"
)

func TestGetBinaryNames(t *testing.T) {
	yt, ff := GetBinaryNames()
	if yt == "" || ff == "" {
		t.Errorf("GetBinaryNames returned empty strings: yt=%s, ff=%s", yt, ff)
	}
}

func TestFindBinaries(t *testing.T) {
	paths, _ := FindBinaries()
	// On this machine, yt-dlp and ffmpeg are in PATH
	if paths.YtDlp == "" {
		t.Logf("Notice: yt-dlp not found in PATH or binDir")
	}
	if paths.FFmpeg == "" {
		t.Logf("Notice: ffmpeg not found in PATH or binDir")
	}
}

func TestGetYtDlpDownloadURL(t *testing.T) {
	url := GetYtDlpDownloadURL()
	if url == "" {
		t.Fatalf("GetYtDlpDownloadURL returned empty URL")
	}
	if !strings.HasPrefix(url, "https://github.com/yt-dlp/yt-dlp/releases/latest/download/") {
		t.Errorf("GetYtDlpDownloadURL returned unexpected URL prefix: %s", url)
	}
}

func TestGetGalleryDlDownloadURL(t *testing.T) {
	url := GetGalleryDlDownloadURL()
	if url == "" {
		t.Fatalf("GetGalleryDlDownloadURL returned empty URL")
	}
	if !strings.HasPrefix(url, "https://codeberg.org/mikf/gallery-dl/releases/download/") {
		t.Errorf("GetGalleryDlDownloadURL returned unexpected URL prefix: %s", url)
	}
}
