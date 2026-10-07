package downloader

import (
	"net/url"
	"strings"
)

// PlatformInfo contains metadata for the media hosting platform.
type PlatformInfo struct {
	Name  string
	Color string
}

func hasDomain(host string, domains ...string) bool {
	for _, d := range domains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// DetectPlatform inspects a URL and returns the platform brand name and style color hex.
func DetectPlatform(rawURL string) PlatformInfo {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return PlatformInfo{Name: "Media", Color: "#05D550"}
	}

	host := strings.ToLower(u.Host)
	switch {
	case hasDomain(host, "youtube.com", "youtu.be"):
		return PlatformInfo{Name: "YouTube", Color: "#FF0000"}
	case hasDomain(host, "tiktok.com", "douyin.com"):
		return PlatformInfo{Name: "TikTok", Color: "#00F2FE"}
	case hasDomain(host, "facebook.com", "fb.watch", "fb.me"):
		return PlatformInfo{Name: "Facebook", Color: "#1877F2"}
	case hasDomain(host, "instagram.com", "instagr.am"):
		return PlatformInfo{Name: "Instagram", Color: "#E1306C"}
	case hasDomain(host, "twitter.com", "x.com", "t.co"):
		return PlatformInfo{Name: "X / Twitter", Color: "#E7E9EA"}
	case hasDomain(host, "soundcloud.com"):
		return PlatformInfo{Name: "SoundCloud", Color: "#FF5500"}
	case hasDomain(host, "spotify.com", "spotify.link"):
		return PlatformInfo{Name: "Spotify", Color: "#1DB954"}
	case hasDomain(host, "reddit.com", "redd.it"):
		return PlatformInfo{Name: "Reddit", Color: "#FF4500"}
	case hasDomain(host, "pinterest.com", "pin.it"):
		return PlatformInfo{Name: "Pinterest", Color: "#E60023"}
	case hasDomain(host, "bilibili.com", "b23.tv"):
		return PlatformInfo{Name: "Bilibili", Color: "#00AEEC"}
	case hasDomain(host, "vimeo.com"):
		return PlatformInfo{Name: "Vimeo", Color: "#1AB7EA"}
	default:
		return PlatformInfo{Name: "Media", Color: "#05D550"}
	}
}
