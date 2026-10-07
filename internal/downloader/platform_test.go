package downloader

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		url          string
		expectedName string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "YouTube"},
		{"https://youtu.be/dQw4w9WgXcQ", "YouTube"},
		{"https://www.tiktok.com/@user/video/12345", "TikTok"},
		{"https://vt.tiktok.com/ZSabc123/", "TikTok"},
		{"https://www.facebook.com/watch/?v=123", "Facebook"},
		{"https://fb.watch/xyz123/", "Facebook"},
		{"https://www.instagram.com/reel/C_abc/", "Instagram"},
		{"https://x.com/user/status/123", "X / Twitter"},
		{"https://soundcloud.com/artist/track", "SoundCloud"},
		{"https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT", "Spotify"},
		{"https://www.reddit.com/r/funny/comments/12345/funny_video/", "Reddit"},
		{"https://www.threads.net/@user/post/abc123", "Threads"},
		{"https://www.pinterest.com/pin/123456789/", "Pinterest"},
		{"https://pin.it/abc123", "Pinterest"},
		{"https://example.com/video.mp4", "Media"},
	}

	for _, tt := range tests {
		p := DetectPlatform(tt.url)
		if p.Name != tt.expectedName {
			t.Errorf("DetectPlatform(%q) = %q; want %q", tt.url, p.Name, tt.expectedName)
		}
	}
}
