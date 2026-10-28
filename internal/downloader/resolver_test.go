package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsRedirectURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://pin.it/abc1234", true},
		{"https://vt.tiktok.com/ZSabc/", true},
		{"https://vm.tiktok.com/ZSabc/", true},
		{"https://www.tiktok.com/t/ZSabc/", true},
		{"https://fb.watch/xyz123/", true},
		{"https://fb.me/xyz123", true},
		{"https://www.facebook.com/share/r/123/", true},
		{"https://www.facebook.com/share/v/123/", true},
		{"https://b23.tv/abc123", true},
		{"https://t.co/xyz123", true},
		{"https://redd.it/12345", true},
		{"https://v.douyin.com/i8xxxx/", true},
		{"https://www.instagram.com/share/p/123/", true},
		{"https://bit.ly/12345", true},
		{"https://spotify.link/abc123", true},
		{"https://youtu.be/dQw4w9WgXcQ", false}, // Handled locally without network request
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"https://www.facebook.com/watch/?v=123", false},
		{"https://example.com/video.mp4", false},
		{"invalid-url", false},
	}

	for _, tt := range tests {
		got := IsRedirectURL(tt.url)
		if got != tt.expected {
			t.Errorf("IsRedirectURL(%q) = %v, want %v", tt.url, got, tt.expected)
		}
	}
}

func TestResolveRedirectMock(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer targetServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL+"/destination", http.StatusFound)
	}))
	defer redirectServer.Close()

	ctx := context.Background()

	// Normal URL that is not in IsRedirectURL should return as-is without network call
	unrelatedURL := redirectServer.URL + "/something"
	if got := ResolveRedirect(ctx, unrelatedURL); got != unrelatedURL {
		t.Errorf("expected %q, got %q", unrelatedURL, got)
	}
}

func TestResolveRedirect_YouTubeShort(t *testing.T) {
	got := ResolveRedirect(context.Background(), "https://youtu.be/dQw4w9WgXcQ?list=PL123&t=10")
	if !strings.Contains(got, "v=dQw4w9WgXcQ") || !strings.Contains(got, "www.youtube.com/watch") {
		t.Errorf("expected youtube.com/watch URL with v=, got: %s", got)
	}
}

func TestIsRedirectURLDoesNotMatchSubstrings(t *testing.T) {
	falsePositives := []string{
		"https://microsoft.com/download",
		"https://videos.id/abc",
		"https://reddit.com/r/pics",
		"https://pin.it.example.com/x",
		"https://spotify.link.evil.com/x",
	}
	for _, u := range falsePositives {
		if IsRedirectURL(u) {
			t.Errorf("IsRedirectURL(%q) = true, want false (substring false positive)", u)
		}
	}

	truePositives := []string{
		"https://t.co/abc123",
		"https://bit.ly/abc123",
		"https://pin.it/abc123",
		"https://vt.tiktok.com/ZS1234/",
		"https://fb.watch/abc123",
		"https://s.id/abc123",
		"https://b23.tv/abc123",
	}
	for _, u := range truePositives {
		if !IsRedirectURL(u) {
			t.Errorf("IsRedirectURL(%q) = false, want true", u)
		}
	}
}
