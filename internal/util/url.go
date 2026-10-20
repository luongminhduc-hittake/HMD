package util

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var spotifyURIRegex = regexp.MustCompile(`^spotify:(track|album|playlist):([a-zA-Z0-9]+)$`)

// SanitizeURL strips quotes, invisible Unicode characters (BOM, zero-width space),
// trims whitespace, auto-prepends https:// for domain inputs, and expands spotify: URIs.
func SanitizeURL(raw string) (string, error) {
	// 1. Remove invisible Unicode runes and whitespace
	cleaned := strings.Map(func(r rune) rune {
		if r == '\uFEFF' || r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\u00A0' {
			return -1
		}
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, raw)
	cleaned = strings.TrimSpace(cleaned)

	// 2. Strip surrounding single or double quotes
	cleaned = strings.Trim(cleaned, "\"'`")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return "", fmt.Errorf("đường dẫn trống")
	}

	// 3. Handle Spotify URI (spotify:track:ID, spotify:album:ID, spotify:playlist:ID)
	if m := spotifyURIRegex.FindStringSubmatch(cleaned); len(m) == 3 {
		return fmt.Sprintf("https://open.spotify.com/%s/%s", m[1], m[2]), nil
	}

	// 4. Auto-prepend https:// if missing scheme
	if !strings.HasPrefix(cleaned, "http://") && !strings.HasPrefix(cleaned, "https://") {
		// If it looks like a domain or URL without scheme
		if strings.Contains(cleaned, ".") && !strings.Contains(cleaned, " ") {
			cleaned = "https://" + cleaned
		} else {
			return "", fmt.Errorf("đường dẫn không hợp lệ. Vui lòng nhập link bắt đầu bằng http:// hoặc https://")
		}
	}

	// 5. Validate with url.Parse
	u, err := url.Parse(cleaned)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("định dạng URL không hợp lệ: %v", err)
	}

	return cleaned, nil
}
