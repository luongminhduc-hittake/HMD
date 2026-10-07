package util

import (
	"testing"
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

