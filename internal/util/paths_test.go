package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1572864, "1.5 MiB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %s; want %s", tt.input, got, tt.expected)
		}
	}
}

func TestDirectories(t *testing.T) {
	binDir, err := GetBinDir()
	if err != nil {
		t.Fatalf("GetBinDir returned error: %v", err)
	}
	if binDir == "" {
		t.Errorf("GetBinDir returned empty string")
	}

	dlDir := GetDefaultDownloadDir()
	if dlDir == "" {
		t.Errorf("GetDefaultDownloadDir returned empty string")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Name", "Normal Name"},
		{"Song: Title / With? Wild* Chars|", "Song_ Title _ With_ Wild_ Chars_"},
		{"...dots...", "dots"},
		{"", "untitled"},
		{"   ", "untitled"},
		{"\x00\x01\x1fHidden\x07", "Hidden"},
	}

	for _, tt := range tests {
		got := SanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTruncateTitlePreservesRunes(t *testing.T) {
	title := "Nguyễn Văn A hát rất hay và đây là một tiêu đề rất dài để thử"

	got := TruncateTitle(title, 50)
	if len([]rune(got)) > 50 {
		t.Errorf("truncated title has %d runes, want <= 50", len([]rune(got)))
	}
	if !utf8.ValidString(got) {
		t.Errorf("truncated title is not valid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
	if !strings.HasPrefix(got, "Nguyễn Văn A") {
		t.Errorf("expected prefix preserved, got %q", got)
	}

	if short := TruncateTitle("Ngắn", 50); short != "Ngắn" {
		t.Errorf("short title must be unchanged, got %q", short)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}

	if got := ExpandHome("~/Videos/hmd"); got != filepath.Join(home, "Videos", "hmd") {
		t.Errorf("ExpandHome(~/Videos/hmd) = %q, want %q", got, filepath.Join(home, "Videos", "hmd"))
	}
	if got := ExpandHome("~"); got != home {
		t.Errorf("ExpandHome(~) = %q, want %q", got, home)
	}
	if got := ExpandHome("/abs/path"); got != "/abs/path" {
		t.Errorf("ExpandHome must not touch absolute paths, got %q", got)
	}
	if got := ExpandHome("rel/path"); got != "rel/path" {
		t.Errorf("ExpandHome must not touch relative paths, got %q", got)
	}
	// "~someone" is another user's home; must stay untouched.
	if got := ExpandHome("~otheruser/x"); got != "~otheruser/x" {
		t.Errorf("ExpandHome must not expand other users, got %q", got)
	}
}
