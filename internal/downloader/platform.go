package downloader

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PlatformInfo contains metadata for the media hosting platform.
type PlatformInfo struct {
	Name  string
	Color lipgloss.Color
}

// DetectPlatform inspects a URL and returns the platform brand name and style color.
func DetectPlatform(rawURL string) PlatformInfo {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return PlatformInfo{Name: "Media", Color: lipgloss.Color("#05D550")}
	}

	host := strings.ToLower(u.Host)
	switch {
	case strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be"):
		return PlatformInfo{Name: "YouTube", Color: lipgloss.Color("#FF0000")}
	case strings.Contains(host, "tiktok.com") || strings.Contains(host, "douyin.com"):
		return PlatformInfo{Name: "TikTok", Color: lipgloss.Color("#00F2FE")}
	case strings.Contains(host, "facebook.com") || strings.Contains(host, "fb.watch"):
		return PlatformInfo{Name: "Facebook", Color: lipgloss.Color("#1877F2")}
	case strings.Contains(host, "instagram.com"):
		return PlatformInfo{Name: "Instagram", Color: lipgloss.Color("#E1306C")}
	case strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com"):
		return PlatformInfo{Name: "X / Twitter", Color: lipgloss.Color("#E7E9EA")}
	case strings.Contains(host, "soundcloud.com"):
		return PlatformInfo{Name: "SoundCloud", Color: lipgloss.Color("#FF5500")}
	case strings.Contains(host, "bilibili.com"):
		return PlatformInfo{Name: "Bilibili", Color: lipgloss.Color("#00AEEC")}
	case strings.Contains(host, "vimeo.com"):
		return PlatformInfo{Name: "Vimeo", Color: lipgloss.Color("#1AB7EA")}
	default:
		return PlatformInfo{Name: "Media", Color: lipgloss.Color("#05D550")}
	}
}
