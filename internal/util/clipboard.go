package util

import (
	"github.com/atotto/clipboard"
)

// GetMediaURLFromClipboard attempts to read the system clipboard and returns
// a sanitized URL if valid.
func GetMediaURLFromClipboard() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	cleanURL, err := SanitizeURL(text)
	if err != nil {
		return ""
	}
	return cleanURL
}
