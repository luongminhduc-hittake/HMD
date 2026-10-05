package util

import (
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
)

// GetMediaURLFromClipboard attempts to read the system clipboard and returns
// a trimmed URL if it starts with http:// or https:// and contains a valid host.
// If reading fails or clipboard does not contain a valid URL, it returns an empty string.
func GetMediaURLFromClipboard() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
		return ""
	}
	u, err := url.Parse(text)
	if err != nil || u.Host == "" {
		return ""
	}
	return text
}
