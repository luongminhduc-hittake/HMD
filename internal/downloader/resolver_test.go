package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://spotify.link/abc123", true},
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
