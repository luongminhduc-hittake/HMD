package deps

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestFindBinariesReportsMissingGalleryDl(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	_, missing := FindBinaries()

	found := false
	for _, m := range missing {
		if m == "gallery-dl" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected gallery-dl to be reported missing when it is not on PATH, got %v", missing)
	}
}

func TestEnsureDependenciesInstallsGalleryDl(t *testing.T) {
	// Provide yt-dlp and ffmpeg so only gallery-dl is missing; on Linux a missing
	// ffmpeg is a fatal "install it yourself" path, which would mask the gallery-dl case.
	pathDir := t.TempDir()
	for _, name := range []string{"yt-dlp", "ffmpeg"} {
		if err := os.WriteFile(filepath.Join(pathDir, name), []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", pathDir)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	var installed []string
	origDownload := downloadBinary
	defer func() { downloadBinary = origDownload }()
	downloadBinary = func(url, destPath, itemName string, cb ProgressCallback, expectedSHA256 string) error {
		installed = append(installed, itemName)
		return os.WriteFile(destPath, []byte("stub"), 0755)
	}

	paths, err := EnsureDependencies(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paths.GalleryDl == "" {
		t.Fatalf("expected EnsureDependencies to install gallery-dl, got empty path")
	}

	sawGallery := false
	for _, name := range installed {
		if name == "gallery-dl" {
			sawGallery = true
		}
	}
	if !sawGallery {
		t.Errorf("expected gallery-dl download attempt, got %v", installed)
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "payload.bin")
	if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// sha256("hello")
	const good = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	if err := verifyFileChecksum(f, good); err != nil {
		t.Errorf("expected checksum match, got %v", err)
	}
	if err := verifyFileChecksum(f, "deadbeef"); err == nil {
		t.Errorf("expected checksum mismatch error")
	}
	if err := verifyFileChecksum(f, ""); err == nil {
		t.Errorf("expected error for missing checksum")
	}
}

func TestDownloadFileRejectsChecksumMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tampered"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "tool")
	err := downloadVerified(srv.URL, dest, "tool", func(name string, dl, total int64, pct float64) {},
		"0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("expected downloadVerified to reject a mismatched payload")
	}
	if _, statErr := os.Stat(dest); statErr == nil {
		t.Errorf("mismatched download must not be left at the destination")
	}
}

func TestParseGalleryDlLatestTag(t *testing.T) {
	api := `[{"tag_name":"v1.32.18","name":"v1.32.18"},{"tag_name":"v1.32.17","name":"v1.32.17"}]`
	got, err := parseLatestTag(api)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v1.32.18" {
		t.Errorf("parseLatestTag = %q, want v1.32.18", got)
	}

	if _, err := parseLatestTag("[]"); err == nil {
		t.Errorf("expected an error for an empty release list")
	}
	if _, err := parseLatestTag("not json"); err == nil {
		t.Errorf("expected an error for malformed JSON")
	}
}

func TestGalleryDlReleaseURLs(t *testing.T) {
	base := galleryDlReleaseBase("v1.32.18")
	if !strings.Contains(base, "v1.32.18") {
		t.Errorf("release base must carry the tag, got %q", base)
	}
	if !strings.HasSuffix(base, "/v1.32.18") {
		t.Errorf("release base must end with the tag, got %q", base)
	}
}

func TestResolveGalleryDlReleaseFallsBackToPinned(t *testing.T) {
	origFetch := fetchText
	defer func() { fetchText = origFetch }()
	fetchText = func(url string) (string, error) {
		return "", errors.New("network down")
	}

	tag, err := resolveGalleryDlTag()
	if err != nil {
		t.Fatalf("expected fallback to the pinned tag, got error %v", err)
	}
	if tag != GalleryDlVersion {
		t.Errorf("fallback tag = %q, want pinned %q", tag, GalleryDlVersion)
	}
}

func TestResolveGalleryDlChecksumUsesPublishedSums(t *testing.T) {
	origFetch := fetchText
	defer func() { fetchText = origFetch }()

	sums := "aaaa  gallery-dl.bin\nbbbb  gallery-dl.exe\n"
	fetchText = func(url string) (string, error) {
		if strings.HasSuffix(url, "SHA256SUMS") {
			return sums, nil
		}
		return "[]", nil
	}

	if got := resolveGalleryDlChecksum("v1.32.18", "gallery-dl.bin"); got != "aaaa" {
		t.Errorf("checksum for gallery-dl.bin = %q, want aaaa", got)
	}
	if got := resolveGalleryDlChecksum("v1.32.18", "gallery-dl.exe"); got != "bbbb" {
		t.Errorf("checksum for gallery-dl.exe = %q, want bbbb", got)
	}
}
