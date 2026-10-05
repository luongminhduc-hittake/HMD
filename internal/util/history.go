package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryEntry records metadata of a successful download.
type HistoryEntry struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Format    string    `json:"format"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

// GetHistoryFilePath returns ~/.config/hmd/history.json
func GetHistoryFilePath() (string, error) {
	cfgFile, err := GetConfigFilePath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(cfgFile)
	return filepath.Join(dir, "history.json"), nil
}

// LoadHistory reads up to 100 recent entries from history.json.
func LoadHistory() []HistoryEntry {
	histFile, err := GetHistoryFilePath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(histFile)
	if err != nil {
		return nil
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// SaveHistory persists entries to history.json.
func SaveHistory(entries []HistoryEntry) error {
	histFile, err := GetHistoryFilePath()
	if err != nil {
		return err
	}
	if len(entries) > 100 {
		entries = entries[:100]
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(histFile, data, 0644)
}

// AddHistoryEntry prepends a new entry and saves history.
func AddHistoryEntry(entry HistoryEntry) error {
	entries := LoadHistory()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	// Filter out identical URL to prevent duplicate entries in list
	var newEntries []HistoryEntry
	newEntries = append(newEntries, entry)
	cleanURL := strings.TrimSpace(entry.URL)
	for _, e := range entries {
		if strings.TrimSpace(e.URL) != cleanURL {
			newEntries = append(newEntries, e)
		}
	}
	return SaveHistory(newEntries)
}

// FindHistoryByURL checks if a URL was previously downloaded.
func FindHistoryByURL(rawURL string) *HistoryEntry {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil
	}
	entries := LoadHistory()
	for i := range entries {
		if strings.TrimSpace(entries[i].URL) == rawURL {
			return &entries[i]
		}
	}
	return nil
}

// DeleteHistoryEntry removes an entry at the given index.
func DeleteHistoryEntry(index int) error {
	entries := LoadHistory()
	if index < 0 || index >= len(entries) {
		return nil
	}
	entries = append(entries[:index], entries[index+1:]...)
	return SaveHistory(entries)
}
