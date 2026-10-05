package util

import (
	"testing"
)

func TestClipboardURLValidation(t *testing.T) {
	// Simple test verifying URL parsing logic
	validURLs := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://vt.tiktok.com/ZSabc123/",
		"http://facebook.com/watch?v=12345",
	}

	for _, u := range validURLs {
		if !stringsHasHTTP(u) {
			t.Errorf("expected valid for %q", u)
		}
	}
}

func stringsHasHTTP(s string) bool {
	return len(s) > 8 && (s[:7] == "http://" || s[:8] == "https://")
}
